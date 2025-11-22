package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/jcroyoaun/totalcompmx/internal/database"
	"github.com/jcroyoaun/totalcompmx/internal/env"
	"github.com/jcroyoaun/totalcompmx/internal/smtp"
	"github.com/jcroyoaun/totalcompmx/internal/version"
	"github.com/jcroyoaun/totalcompmx/internal/worker"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/lmittmann/tint"
)

func init() {
	// Register types for gob encoding (used by session manager)
	gob.Register([]PackageInput{})
	gob.Register(PackageInput{})
	gob.Register([]PackageResult{})
	gob.Register(PackageResult{})
	gob.Register([]OtherBenefit{})
	gob.Register(OtherBenefit{})
	gob.Register(database.SalaryCalculation{})
	gob.Register(database.FiscalYear{})
	gob.Register([]database.OtherBenefitResult{})
	gob.Register(database.OtherBenefitResult{})
}

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL   string
	httpPort  int
	basicAuth struct {
		username       string
		hashedPassword string
	}
	autoHTTPS struct {
		domain  string
		email   string
		staging bool
	}
	cookie struct {
		secretKey string
	}
	db struct {
		dsn         string
		automigrate bool
	}
	notifications struct {
		email string
	}
	session struct {
		cookieName string
	}
	resend struct {
		apiKey string
		from   string
	}
	worker struct {
		banxicoToken string
		inegiToken   string
	}
}

type application struct {
	config         config
	db             *database.DB
	logger         *slog.Logger
	mailer         *smtp.Mailer
	sessionManager *scs.SessionManager
	wg             sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:3080")
	cfg.httpPort = env.GetInt("HTTP_PORT", 3080)
	cfg.autoHTTPS.domain = env.GetString("AUTO_HTTPS_DOMAIN", "")
	cfg.autoHTTPS.email = env.GetString("AUTO_HTTPS_EMAIL", "admin@github.com/jcroyoaun/totalcompmx")
	cfg.autoHTTPS.staging = env.GetBool("AUTO_HTTPS_STAGING", false)
	cfg.basicAuth.username = env.GetString("BASIC_AUTH_USERNAME", "admin")
	cfg.basicAuth.hashedPassword = env.GetString("BASIC_AUTH_HASHED_PASSWORD", "$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa")
	cfg.cookie.secretKey = env.GetString("COOKIE_SECRET_KEY", "v2zn7or6otz36wqjt7b2qkj2xj3g7ug5")
	cfg.db.dsn = env.GetString("DB_DSN", "user:pass@localhost:5432/db")
	cfg.db.automigrate = env.GetBool("DB_AUTOMIGRATE", true)
	cfg.notifications.email = env.GetString("NOTIFICATIONS_EMAIL", "")
	cfg.session.cookieName = env.GetString("SESSION_COOKIE_NAME", "session_uprn7vcq")
	cfg.resend.apiKey = env.GetString("RESEND_API_KEY", "")
	cfg.resend.from = env.GetString("RESEND_FROM", "TotalComp MX <hola@totalcomp.mx>")
	cfg.worker.banxicoToken = env.GetString("BANXICO_TOKEN", "")
	cfg.worker.inegiToken = env.GetString("INEGI_TOKEN", "")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if cfg.db.automigrate {
		err = db.MigrateUp()
		if err != nil {
			return err
		}
	}

	// Initialize mailer with Resend
	// If no API key is provided, use mock mailer for testing
	var mailer *smtp.Mailer
	if cfg.resend.apiKey == "" {
		logger.Warn("RESEND_API_KEY not set, using mock mailer (emails will not be sent)")
		mailer = smtp.NewMockMailer(cfg.resend.from)
	} else {
		mailer = smtp.NewMailer(cfg.resend.apiKey, cfg.resend.from)
	}

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db.DB.DB)
	sessionManager.Lifetime = 7 * 24 * time.Hour
	sessionManager.Cookie.Name = cfg.session.cookieName
	sessionManager.Cookie.Secure = true

	app := &application{
		config:         cfg,
		db:             db,
		logger:         logger,
		mailer:         mailer,
		sessionManager: sessionManager,
	}

	// Start background ETL worker for financial data updates
	etlWorker := worker.New(db, logger, cfg.worker.banxicoToken, cfg.worker.inegiToken)
	go etlWorker.Start()
	logger.Info("etl worker started", "banxico_enabled", true, "inegi_enabled", true)

	if cfg.autoHTTPS.domain != "" {
		return app.serveAutoHTTPS()
	}

	return app.serveHTTP()
}
