package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"strconv"
	"time"

	"github.com/jcroyoaun/totalcompmx/internal/metrics"
	"github.com/jcroyoaun/totalcompmx/internal/response"

	"github.com/justinas/nosurf"
	"github.com/tomasen/realip"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pv := recover()
			if pv != nil {
				app.serverError(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := response.NewMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		var (
			ip     = realip.FromRequest(r)
			method = r.Method
			url    = r.URL.String()
			proto  = r.Proto
		)

		userAttrs := slog.Group("user", "ip", ip)
		requestAttrs := slog.Group("request", "method", method, "url", url, "proto", proto)
		responseAttrs := slog.Group("response", "status", mw.StatusCode, "size", mw.BytesCount)

		app.logger.Info("access", userAttrs, requestAttrs, responseAttrs)
	})
}

func (app *application) preventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.badRequest(w, r, errors.New("CSRF token validation failed"))
	}))

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		user, found, err := app.db.GetUser(id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		if found {
			r = contextSetAuthenticatedUser(r, user)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := contextGetAuthenticatedUser(r)
		if !found {
			app.sessionManager.Put(r.Context(), "redirectPathAfterLogin", r.URL.Path)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAnonymousUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := contextGetAuthenticatedUser(r)

		if found {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return

		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireBasicAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, plaintextPassword, ok := r.BasicAuth()
		if !ok {
			app.basicAuthenticationRequired(w, r)
			return
		}

		if app.config.basicAuth.username != username {
			app.basicAuthenticationRequired(w, r)
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(app.config.basicAuth.hashedPassword), []byte(plaintextPassword))
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			app.basicAuthenticationRequired(w, r)
			return
		case err != nil:
			app.serverError(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requireAPIKey validates the API key in the Authorization header (stateless)
func (app *application) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := response.JSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Missing Authorization header. Use: Authorization: Bearer YOUR_API_KEY",
			})
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		// Expect format: "Bearer <API_KEY>"
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			err := response.JSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid Authorization format. Use: Authorization: Bearer YOUR_API_KEY",
			})
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		apiKey := authHeader[len(bearerPrefix):]
		if apiKey == "" {
			err := response.JSON(w, http.StatusUnauthorized, map[string]string{
				"error": "API key is empty",
			})
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		// Look up user by API key
		user, found, err := app.db.GetUserByAPIKey(apiKey)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		if !found {
			err := response.JSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid API key",
			})
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		// RATE LIMITING: Check if user has exceeded their limit
		// Unverified users: 10 calls/day
		// Verified users: 100 calls/month (TODO: implement monthly check)
		if !user.EmailVerified {
			// Check daily limit for unverified users
			dailyCount, err := app.db.GetDailyAPICallCount(user.ID)
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			if dailyCount >= 10 {
				err := response.JSON(w, http.StatusTooManyRequests, map[string]interface{}{
					"error":   "Daily API limit exceeded",
					"message": "You have reached your daily limit of 10 API calls. Verify your email to unlock 100 calls/month.",
					"limit":   10,
					"used":    dailyCount,
					"type":    "unverified_user",
					"action":  "Please verify your email to increase your limit.",
				})
				if err != nil {
					app.serverError(w, r, err)
				}
				return
			}
		}

		// Log the API call for rate limiting
		err = app.db.LogAPICall(user.ID)
		if err != nil {
			// Don't block the request if logging fails, but log the error
			app.logger.Error("failed to log API call", "error", err, "user_id", user.ID)
		}

		// Increment API calls counter (fire and forget, don't block on errors)
		go func() {
			_ = app.db.IncrementAPICallsCount(user.ID)
		}()

		// Store user in request context for handler access
		r = contextSetAuthenticatedUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *application) prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip metrics endpoint to avoid pollution
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Use the custom response writer to capture status code
		// Check if w is already a MetricsResponseWriter (e.g. from logAccess)
		var mw *response.MetricsResponseWriter
		if v, ok := w.(*response.MetricsResponseWriter); ok {
			mw = v
		} else {
			mw = response.NewMetricsResponseWriter(w)
		}

		// Increment active requests
		metrics.ActiveRequests.Inc()
		defer metrics.ActiveRequests.Dec()

		// If we wrapped it, use mw, otherwise w is already mw
		next.ServeHTTP(mw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(mw.StatusCode)

		// Record metrics
		metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
		metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
	})
}
