package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sanchey92/metric-agent/internal/models"
)

func TestSender_Run(t *testing.T) {
	type call struct {
		statusCode int
		bodyCheck  func([]models.Metric, *testing.T)
	}

	tests := []struct {
		name           string
		interval       time.Duration
		waitBeforeStop time.Duration
		metricsToSend  []models.Metric
		expectedCalls  int
		responses      []call
	}{
		{
			name:     "sends metrics by interval",
			interval: 100 * time.Millisecond,
			metricsToSend: []models.Metric{
				{Name: "Alloc", MType: "gauge", Value: 123},
				{Name: "HeapSys", MType: "gauge", Value: 456},
			},
			waitBeforeStop: 150 * time.Millisecond,
			expectedCalls:  1,
			responses: []call{
				{
					statusCode: http.StatusOK,
					bodyCheck: func(m []models.Metric, t *testing.T) {
						assert.Len(t, m, 2)
					},
				},
			},
		},
		{
			name:     "flushes on ctx.Done",
			interval: 10 * time.Second,
			metricsToSend: []models.Metric{
				{Name: "Alloc", MType: "gauge", Value: 123},
			},
			waitBeforeStop: 50 * time.Millisecond,
			expectedCalls:  1,
			responses: []call{
				{
					statusCode: http.StatusOK,
					bodyCheck: func(m []models.Metric, t *testing.T) {
						assert.Equal(t, "Alloc", m[0].Name)
					},
				},
			},
		},
		{
			name:           "no metrics, no send",
			interval:       50 * time.Millisecond,
			waitBeforeStop: 100 * time.Millisecond,
			expectedCalls:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close() //nolint:errcheck

				var buf bytes.Buffer
				gr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)

				defer gr.Close() //nolint:errcheck

				limitedReader := io.LimitReader(gr, 1<<20)
				_, err = io.Copy(&buf, limitedReader)
				require.NoError(t, err)

				var received []models.Metric
				require.NoError(t, json.Unmarshal(buf.Bytes(), &received))

				if tt.responses != nil && callCount < len(tt.responses) {
					resp := tt.responses[callCount]
					if resp.bodyCheck != nil {
						resp.bodyCheck(received, t)
					}
					w.WriteHeader(resp.statusCode)
				}
				callCount++
			}))
			defer server.Close()

			s := New(server.URL, tt.interval)

			metricsCh := make(chan models.Metric, len(tt.metricsToSend))
			for _, m := range tt.metricsToSend {
				metricsCh <- m
			}
			close(metricsCh)

			ctx, cancel := context.WithTimeout(context.Background(), tt.waitBeforeStop)
			defer cancel()

			err := s.Run(ctx, metricsCh)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCalls, callCount)
		})
	}
}
