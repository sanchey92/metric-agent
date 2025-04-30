// Package sender provides functionality for sending metrics to a remote endpoint
package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sanchey92/metric-agent/internal/models"
)

// Sender sends metrics to a remote endpoint in batches.
type Sender struct {
	reportInterval time.Duration
	endpoint       string
	client         *http.Client
}

// New creates a new Sender with the specified endpoint and report interval.
func New(endpoint string, interval time.Duration) *Sender {
	return &Sender{
		reportInterval: interval,
		endpoint:       endpoint,
		client:         &http.Client{Timeout: 5 * time.Second},
	}
}

// Run processes metrics from the channel and sends them in batches.
func (s *Sender) Run(ctx context.Context, metricsCh <-chan models.Metric) error {
	ticker := time.NewTicker(s.reportInterval)
	defer ticker.Stop()

	buffer := make([]models.Metric, 0, 100)

	for {
		select {
		case <-ctx.Done():
			if len(buffer) > 0 {
				if err := s.sendBatch(ctx, buffer); err != nil {
					log.Printf("[sender] failed to send final batch: %v", err)
				}
			}
			return ctx.Err()

		case metric, ok := <-metricsCh:
			if !ok {
				if len(buffer) > 0 {
					if err := s.sendBatch(ctx, buffer); err != nil {
						log.Printf("[sender] failed to send batch after channel close: %v", err)
					}
				}
				return nil
			}
			buffer = append(buffer, metric)

		case <-ticker.C:
			if len(buffer) == 0 {
				continue
			}

			if err := s.sendBatch(ctx, buffer); err != nil {
				log.Printf("[sender] failed to send batch: %v", err)
			}
			buffer = buffer[:0]
		}
	}
}

func (s *Sender) sendBatch(ctx context.Context, metrics []models.Metric) error {
	jsonData, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err = gz.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write gzip data: %w", err)
	}

	if err := gz.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully sent batch of %d metrics to %s", len(metrics), s.endpoint)
	return nil
}
