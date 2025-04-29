// Package collector provides functionality for collecting runtime metrics
// from the Go application at specified intervals.
package collector

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/sanchey92/metric-agent/internal/models"
)

// Collector represents a runtime metrics collector that periodically
// gathers system metrics and sends them through a channel.
type Collector struct {
	pollInterval time.Duration
	pollCount    int64
}

// New creates and returns a new Collector instance with the specified polling interval.
func New(interval time.Duration) *Collector {
	return &Collector{
		pollInterval: interval,
		pollCount:    0,
	}
}

// Run starts the metric collection process. It periodically collects metrics
// and sends them to the provided channel. This method blocks until the context
// is canceled or an error occurs during metric collection or transmission.
func (c *Collector) Run(ctx context.Context, metricsCh chan<- models.Metric) error {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.sendToChannel(ctx, metricsCh); err != nil {
				return err
			}
			atomic.AddInt64(&c.pollCount, 1)
		}
	}
}

func (c *Collector) sendToChannel(ctx context.Context, metricsCh chan<- models.Metric) error {
	metrics := collectMetrics()

	for _, m := range metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case metricsCh <- m:
		}
	}

	return nil
}

func collectMetrics() []models.Metric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := []models.Metric{
		newGauge("Alloc", float64(m.Alloc)),
		newGauge("BuckHashSys", float64(m.BuckHashSys)),
		newGauge("Frees", float64(m.Frees)),
		newGauge("GCCPUFraction", m.GCCPUFraction),
		newGauge("GCSys", float64(m.GCSys)),
		newGauge("HeapAlloc", float64(m.HeapAlloc)),
		newGauge("HeapIdle", float64(m.HeapIdle)),
		newGauge("HeapInuse", float64(m.HeapInuse)),
		newGauge("HeapObjects", float64(m.HeapObjects)),
		newGauge("HeapReleased", float64(m.HeapReleased)),
		newGauge("HeapSys", float64(m.HeapSys)),
		newGauge("LastGC", float64(m.LastGC)),
		newGauge("Lookups", float64(m.Lookups)),
		newGauge("MCacheInuse", float64(m.MCacheInuse)),
		newGauge("MCacheSys", float64(m.MCacheSys)),
		newGauge("MSpanInuse", float64(m.MSpanInuse)),
		newGauge("MSpanSys", float64(m.MSpanSys)),
		newGauge("Mallocs", float64(m.Mallocs)),
		newGauge("NextGC", float64(m.NextGC)),
		newGauge("NumForcedGC", float64(m.NumForcedGC)),
		newGauge("NumGC", float64(m.NumGC)),
		newGauge("OtherSys", float64(m.OtherSys)),
		newGauge("PauseTotalNs", float64(m.PauseTotalNs)),
		newGauge("StackInuse", float64(m.StackInuse)),
		newGauge("StackSys", float64(m.StackSys)),
		newGauge("Sys", float64(m.Sys)),
		newGauge("TotalAlloc", float64(m.TotalAlloc)),
	}

	return metrics
}

func newGauge(name string, value float64) models.Metric {
	return models.Metric{
		MType: "gauge",
		Name:  name,
		Value: value,
	}
}
