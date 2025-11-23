package database

import (
	"context"
	"errors"
	"time"

	"github.com/jcroyoaun/totalcompmx/assets"
	"github.com/jcroyoaun/totalcompmx/internal/metrics"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

const defaultTimeout = 3 * time.Second

type DB struct {
	dsn string
	*sqlx.DB
}

func New(dsn string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", "postgres://"+dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(2 * time.Hour)

	return &DB{dsn: dsn, DB: db}, nil
}

func (db *DB) MigrateUp() error {
	iofsDriver, err := iofs.New(assets.EmbeddedFiles, "migrations")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", iofsDriver, "postgres://"+db.dsn)
	if err != nil {
		return err
	}

	err = migrator.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		return nil
	default:
		return err
	}
}

func (db *DB) MonitorConnectionPool() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			// 1. Record Stats (Existing)
			stats := db.DB.Stats()
			metrics.DbOpenConnections.Set(float64(stats.OpenConnections))
			metrics.DbInUseConnections.Set(float64(stats.InUse))
			metrics.DbIdleConnections.Set(float64(stats.Idle))

			// 2. Record Uptime (NEW - The "Ping")
			// We use a short timeout because if it takes >1s, it's effectively down for a user
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			if err := db.PingContext(ctx); err != nil {
				metrics.DbUp.Set(0) // Down
			} else {
				metrics.DbUp.Set(1) // Up
			}
			cancel()
		}
	}()
}
