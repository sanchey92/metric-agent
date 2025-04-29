package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sanchey92/metric-agent/internal/models"
)

func TestCollector_Run(t *testing.T) {
	tests := []struct {
		name           string
		pollInterval   time.Duration
		ctxTimeout     time.Duration
		waitDuration   time.Duration
		expectMinCount int
	}{
		{
			name:           "Should collect metrics once",
			pollInterval:   50 * time.Millisecond,
			ctxTimeout:     200 * time.Millisecond,
			waitDuration:   60 * time.Millisecond,
			expectMinCount: 1,
		},
		{
			name:           "Should collect metrics multiple times",
			pollInterval:   50 * time.Millisecond,
			ctxTimeout:     200 * time.Millisecond,
			waitDuration:   160 * time.Millisecond,
			expectMinCount: 2,
		},
		{
			name:           "Should stop on context cancel",
			pollInterval:   1 * time.Hour,
			ctxTimeout:     50 * time.Millisecond,
			waitDuration:   10 * time.Millisecond,
			expectMinCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.pollInterval)
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()

			metricsCh := make(chan models.Metric, 100)
			errCh := make(chan error, 1)

			go func() {
				errCh <- c.Run(ctx, metricsCh)
			}()

			time.Sleep(tt.waitDuration)

			collected := drainMetrics(metricsCh)

			err := <-errCh
			close(metricsCh)

			assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded),
				"Expected context cancellation error")

			assert.GreaterOrEqual(t, len(collected), tt.expectMinCount,
				"Expected at least %d collected metrics, got %d", tt.expectMinCount, len(collected))
		})
	}
}

// drainMetrics считывает все доступные метрики из канала.
func drainMetrics(ch <-chan models.Metric) []models.Metric {
	var result []models.Metric
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				return result
			}
			result = append(result, m)
		default:
			return result
		}
	}
}
