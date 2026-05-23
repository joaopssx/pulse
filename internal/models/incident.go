package models

import "time"

type Incident struct {
	ID          string         `json:"id" yaml:"id"`
	ServiceID   string         `json:"service_id" yaml:"service_id"`
	ServiceName string         `json:"service_name" yaml:"service_name"`
	StartedAt   time.Time      `json:"started_at" yaml:"started_at"`
	ResolvedAt  *time.Time     `json:"resolved_at" yaml:"resolved_at"`
	Duration    *time.Duration `json:"duration" yaml:"duration"`
	Cause       string         `json:"cause" yaml:"cause"`
	StatusCode  int            `json:"status_code" yaml:"status_code"`
	LatencyMs   int64          `json:"latency_ms" yaml:"latency_ms"`
	Resolved    bool           `json:"resolved" yaml:"resolved"`
}
