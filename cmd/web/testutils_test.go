package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jcroyoaun/totalcompmx/internal/database"
	"github.com/jcroyoaun/totalcompmx/internal/smtp"

	"github.com/alexedwards/scs/v2"
	"github.com/andybalholm/cascadia"
	"github.com/justinas/nosurf"
	"golang.org/x/net/html"
)

type testUser struct {
	id             int
	email          string
	password       string
	hashedPassword string
}

var testUsers = map[string]*testUser{
	"alice": {email: "alice@github.com/jcroyoaun/totalcompmx", password: "testPass123!", hashedPassword: "$2a$04$27fHaQw5jwiMKYoxhLek4uyj9zp29lxtmLWGuC0MR6tuispXJn9US"},
	"bob":   {email: "bob@github.com/jcroyoaun/totalcompmx", password: "mySecure456#", hashedPassword: "$2a$04$O6QOPBSFw14SyLBXs64MJuQd8o7GaBKYvbDqeHGgZRi6FN87aXDWC"},
}

func newTestApplication(t *testing.T) *application {
	app := new(application)

	app.config.session.cookieName = "session_test"

	app.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	app.db = newTestDB(t)
	app.mailer = smtp.NewMockMailer("test@github.com/jcroyoaun/totalcompmx")

	app.sessionManager = scs.New()
	app.sessionManager.Lifetime = 7 * 24 * time.Hour
	app.sessionManager.Cookie.Name = app.config.session.cookieName
	app.sessionManager.Cookie.Secure = true

	return app
}

func newTestDB(t *testing.T) *database.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Fatal("TEST_DB_DSN environment variable must be set in the format user:pass@localhost:port/db")
	}

	schemaName := fmt.Sprintf("test_schema_%d", time.Now().UnixNano())
	dsn = fmt.Sprintf("%s?search_path=%s", dsn, schemaName)

	db, err := database.New(dsn)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
		if err != nil {
			t.Error(err)
		}
	})

	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	if err != nil {
		t.Fatal(err)
	}

	err = db.MigrateUp()
	if err != nil {
		t.Fatal(err)
	}

	for _, user := range testUsers {
		id, err := db.InsertUser(user.email, user.hashedPassword)
		if err != nil {
			t.Fatal(err)
		}

		user.id = id
	}

	return db
}

func newTestRequest(t *testing.T, method, path string) *http.Request {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Form = url.Values{}
	req.PostForm = url.Values{}

	req.Header.Set("Sec-Fetch-Site", "same-origin")
	return req
}

type testResponse struct {
	*http.Response
	Body string
}

func send(t *testing.T, req *http.Request, h http.Handler) testResponse {
	if len(req.PostForm) > 0 {
		body := req.PostForm.Encode()
		req.Body = io.NopCloser(strings.NewReader(body))
		req.ContentLength = int64(len(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return testResponse{
		Response: res,
		Body:     strings.TrimSpace(string(resBody)),
	}
}

func sendWithCSRFToken(t *testing.T, req *http.Request, h http.Handler) testResponse {
	csrfToken, csrfCookie := getValidCSRFData(t)
	req.AddCookie(csrfCookie)
	req.PostForm.Set("csrf_token", csrfToken)

	return send(t, req, h)
}

type testSession struct {
	token  string
	cookie *http.Cookie
	data   map[string]any
}

func newTestSession(t *testing.T, sessionManager *scs.SessionManager, data map[string]any) testSession {
	ctx, err := sessionManager.Load(t.Context(), "")
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range data {
		sessionManager.Put(ctx, key, value)
	}

	sessionToken, _, err := sessionManager.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	sessionCookie := &http.Cookie{
		Name:  sessionManager.Cookie.Name,
		Value: sessionToken,
	}

	return testSession{
		token:  sessionToken,
		cookie: sessionCookie,
		data:   data,
	}
}

func getTestSession(t *testing.T, sessionManager *scs.SessionManager, responseCookies []*http.Cookie) *testSession {
	session := testSession{
		data: make(map[string]any),
	}

	for _, cookie := range responseCookies {
		if cookie.Name == sessionManager.Cookie.Name {
			session.token = cookie.Value
			session.cookie = cookie

			ctx, err := sessionManager.Load(t.Context(), session.token)
			if err != nil {
				t.Fatal(err)
			}

			for _, key := range sessionManager.Keys(ctx) {
				session.data[key] = sessionManager.Get(ctx, key)
			}

			return &session
		}
	}

	return nil
}

func containsPageTag(t *testing.T, htmlBody string, tag string) bool {
	return containsHTMLNode(t, htmlBody, fmt.Sprintf(`meta[name="page"][content="%s"]`, tag))
}

func containsHTMLNode(t *testing.T, htmlBody string, cssSelector string) bool {
	_, found := getHTMLNode(t, htmlBody, cssSelector)
	return found
}

func getHTMLNode(t *testing.T, htmlBody string, cssSelector string) (*html.Node, bool) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		t.Fatal(err)
	}

	selector, err := cascadia.Compile(cssSelector)
	if err != nil {
		t.Fatal(err)
	}

	node := cascadia.Query(doc, selector)
	if node == nil {
		return nil, false
	}

	return node, true
}

func getValidCSRFData(t *testing.T) (string, *http.Cookie) {
	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	var (
		csrfToken  string
		csrfCookie *http.Cookie
	)

	nosurf.NewPure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfToken = nosurf.Token(r)
	})).ServeHTTP(res, req)

	for _, ck := range res.Result().Cookies() {
		if ck.Name == "csrf_token" {
			csrfCookie = ck
			break
		}
	}

	if !nosurf.VerifyToken(csrfToken, csrfCookie.Value) {
		t.Fatalf("unable to generate CSRF token and cookie")
	}

	return csrfToken, csrfCookie
}
