package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pulse/pulse/internal/models"
)

type DispatchError struct {
	ServiceName  string
	AttemptCount int
	Err          error
}

func (e DispatchError) Error() string {
	return fmt.Sprintf("failed to send alert for %s after %d attempts: %v", e.ServiceName, e.AttemptCount, e.Err)
}

type BotDispatcher struct {
	endpoint string
	secret   string
	timeout  time.Duration
	client   *http.Client
}

func NewBotDispatcher(endpoint, secret string, timeout time.Duration) *BotDispatcher {
	return &BotDispatcher{
		endpoint: endpoint,
		secret:   secret,
		timeout:  timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (b *BotDispatcher) Send(ctx context.Context, alert models.Alert) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return DispatchError{ServiceName: alert.ServiceName, AttemptCount: 0, Err: err}
	}

	url := b.endpoint + "/alert"
	backoffs := []time.Duration{
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
	}
	attempts := 0

	var lastErr error
	for {
		attempts++

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return DispatchError{ServiceName: alert.ServiceName, AttemptCount: attempts, Err: err}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Pulse-Secret", b.secret)

		resp, err := b.client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempts > len(backoffs) {
			break
		}

		select {
		case <-ctx.Done():
			return DispatchError{ServiceName: alert.ServiceName, AttemptCount: attempts, Err: ctx.Err()}
		case <-time.After(backoffs[attempts-1]):
		}
	}

	return DispatchError{ServiceName: alert.ServiceName, AttemptCount: attempts, Err: lastErr}
}
