package models

import "time"

type ServiceStatus string

const (
	ServiceStatusHealthy  ServiceStatus = "healthy"
	ServiceStatusDegraded ServiceStatus = "degraded"
	ServiceStatusDown     ServiceStatus = "down"
	ServiceStatusUnknown  ServiceStatus = "unknown"
)

type Service struct {
	ID             string        `json:"id" yaml:"id"`
	Name           string        `json:"name" yaml:"name"`
	URL            string        `json:"url" yaml:"url"`
	Interval       time.Duration `json:"interval" yaml:"interval"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	ExpectedStatus int           `json:"expected_status" yaml:"expected_status"`
	Tags           []string      `json:"tags" yaml:"tags"`
	AlertChannels  []string      `json:"alert_channels" yaml:"alert_channels"`
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	CreatedAt      time.Time     `json:"created_at" yaml:"created_at"`
}
