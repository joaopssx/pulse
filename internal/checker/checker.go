package checker

import (
	"context"
	"net/http"
	"time"

	"github.com/pulse/pulse/internal/models"
)

type Checker struct {
	service models.Service
	client  *http.Client
}

func NewChecker(service models.Service) *Checker {
	return &Checker{
		service: service,
		client: &http.Client{
			Timeout: service.Timeout,
		},
	}
}

func (c *Checker) Check(ctx context.Context) Result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.service.URL, nil)

	start := time.Now()
	var statusCode int
	var success bool
	var errStr string

	if err != nil {
		errStr = err.Error()
	} else {
		resp, err := c.client.Do(req)
		if err != nil {
			errStr = err.Error()
		} else {
			defer resp.Body.Close()
			statusCode = resp.StatusCode
			if statusCode == c.service.ExpectedStatus {
				success = true
			}
		}
	}

	latency := time.Since(start).Milliseconds()

	return Result{
		ServiceID:   c.service.ID,
		ServiceName: c.service.Name,
		URL:         c.service.URL,
		CheckedAt:   start,
		StatusCode:  statusCode,
		LatencyMs:   latency,
		Success:     success,
		Error:       errStr,
	}
}
