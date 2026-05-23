package checker

import "time"

type Result struct {
	ID          string
	ServiceID   string
	ServiceName string
	URL         string
	CheckedAt   time.Time
	StatusCode  int
	LatencyMs   int64
	Success     bool
	Error       string
}
