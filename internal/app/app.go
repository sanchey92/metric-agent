// Package app ties together the collector and sender components,
// orchestrating metric collection and delivery with graceful shutdown.
package app

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/sanchey92/metric-agent/internal/collector"
	"github.com/sanchey92/metric-agent/internal/config"
	"github.com/sanchey92/metric-agent/internal/models"
	"github.com/sanchey92/metric-agent/internal/sender"
)

// App encapsulates configuration, channels, and error handling
// for running the metric-collection agent.
type App struct {
	cfg      *config.Config
	metricCh chan models.Metric
	errCh    chan error
}

// New constructs an App instance with the given configuration.
// It initializes internal channels for metrics and error propagation.
func New(cfg *config.Config) *App {
	return &App{
		cfg:      cfg,
		metricCh: make(chan models.Metric, 100),
		errCh:    make(chan error, 2),
	}
}

// Run starts the collector and sender in separate goroutines,
// listens for OS interrupt or termination signals, and returns
// when either a critical error occurs or a shutdown signal is received.
func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c := collector.New(a.cfg.PollInterval)
	s := sender.New(a.cfg.ServerAddr, a.cfg.ReportInterval)

	go func() {
		defer close(a.metricCh)
		log.Println("Collector started")
		if err := c.Run(ctx, a.metricCh); err != nil {
			a.errCh <- err
		}
	}()

	go func() {
		log.Println("Sender started")
		if err := s.Run(ctx, a.metricCh); err != nil {
			a.errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Agent shutdown initiated")
		return nil
	case err := <-a.errCh:
		return err
	}
}
