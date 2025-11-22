package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jcroyoaun/totalcompmx/internal/database"
)

// Worker manages background ETL jobs for financial data updates
type Worker struct {
	db     *database.DB
	logger *slog.Logger
	
	// Clients
	banxico *BanxicoClient
	inegi   *INEGIClient
	
	// Control
	stopChan chan struct{}
}

// New creates a new Worker instance
func New(db *database.DB, logger *slog.Logger, banxicoToken, inegiToken string) *Worker {
	return &Worker{
		db:       db,
		logger:   logger,
		banxico:  NewBanxicoClient(banxicoToken, logger),
		inegi:    NewINEGIClient(inegiToken, logger),
		stopChan: make(chan struct{}),
	}
}

// Start begins the worker's job scheduling
func (w *Worker) Start() {
	w.logger.Info("worker started", "service", "etl_worker")

	// Run initial jobs immediately on startup
	w.runInitialJobs()

	// Schedule Banxico (Daily at 14:00 CST)
	go w.scheduleBanxico()

	// Schedule INEGI (Weekly on Mondays at 09:00 CST)
	go w.scheduleINEGI()

	// Wait for stop signal
	<-w.stopChan
	w.logger.Info("worker stopped", "service", "etl_worker")
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop() {
	close(w.stopChan)
}

// runInitialJobs executes all jobs once at startup
func (w *Worker) runInitialJobs() {
	w.logger.Info("running initial ETL jobs")
	
	// Run Banxico update (skip if no token configured)
	if w.banxico.token == "" {
		w.logger.Warn("skipping initial Banxico update - no token configured")
	} else {
		if err := w.updateExchangeRate(); err != nil {
			w.logger.Error("initial banxico update failed", "error", err)
		}
	}
	
	// Run INEGI update (skip if no token configured)
	if w.inegi.token == "" {
		w.logger.Warn("skipping initial INEGI update - no token configured")
	} else {
		if err := w.updateUMA(); err != nil {
			w.logger.Error("initial inegi update failed", "error", err)
		}
	}
}

// scheduleBanxico runs daily at 14:00 CST
func (w *Worker) scheduleBanxico() {
	// Skip scheduling if no token configured
	if w.banxico.token == "" {
		w.logger.Info("banxico scheduler disabled - no token configured")
		return
	}

	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().In(cstLocation())
			
			// Run at 14:00 CST daily
			if now.Hour() == 14 && now.Minute() < 60 {
				w.logger.Info("running scheduled banxico update")
				if err := w.updateExchangeRate(); err != nil {
					w.logger.Error("banxico update failed", "error", err)
				}
			}
		case <-w.stopChan:
			return
		}
	}
}

// scheduleINEGI runs weekly on Mondays at 09:00 CST
func (w *Worker) scheduleINEGI() {
	// Skip scheduling if no token configured
	if w.inegi.token == "" {
		w.logger.Info("inegi scheduler disabled - no token configured")
		return
	}

	ticker := time.NewTicker(6 * time.Hour) // Check every 6 hours
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().In(cstLocation())
			
			// Run on Mondays at 09:00 CST
			if now.Weekday() == time.Monday && now.Hour() == 9 && now.Minute() < 360 {
				w.logger.Info("running scheduled inegi update")
				if err := w.updateUMA(); err != nil {
					w.logger.Error("inegi update failed", "error", err)
				}
			}
		case <-w.stopChan:
			return
		}
	}
}

// updateExchangeRate fetches and updates USD/MXN rate from Banxico
func (w *Worker) updateExchangeRate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rate, err := w.banxico.GetExchangeRate(ctx)
	if err != nil {
		return err
	}

	// Update the active fiscal year
	if err := w.db.UpdateExchangeRate(rate); err != nil {
		return err
	}

	w.logger.Info("exchange rate updated", "rate", rate, "source", "banxico")
	return nil
}

// updateUMA fetches and updates UMA from INEGI
func (w *Worker) updateUMA() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	umaData, err := w.inegi.GetUMA(ctx)
	if err != nil {
		return err
	}

	// Update the active fiscal year with new UMA values
	if err := w.db.UpdateUMA(umaData.Annual, umaData.Monthly, umaData.Daily); err != nil {
		return err
	}

	w.logger.Info("uma updated", 
		"annual", umaData.Annual, 
		"monthly", umaData.Monthly, 
		"daily", umaData.Daily,
		"source", "inegi")
	return nil
}

// cstLocation returns CST timezone
func cstLocation() *time.Location {
	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		// Fallback to UTC-6 (CST)
		return time.FixedZone("CST", -6*60*60)
	}
	return loc
}

