package models

import "time"

type AlertType string

const (
	AlertTypeDown        AlertType = "down"
	AlertTypeSlow        AlertType = "slow"
	AlertTypeRecovered   AlertType = "recovered"
	AlertTypeSSLExpiring AlertType = "ssl_expiring"
	AlertTypeAnomaly     AlertType = "anomaly"
)

type Alert struct {
	ID              string    `json:"id" yaml:"id"`
	IncidentID      string    `json:"incident_id" yaml:"incident_id"`
	ServiceName     string    `json:"service_name" yaml:"service_name"`
	Status          string    `json:"status" yaml:"status"`
	URL             string    `json:"url" yaml:"url"`
	LatencyMs       int64     `json:"latency_ms" yaml:"latency_ms"`
	NormalLatencyMs int64     `json:"normal_latency_ms" yaml:"normal_latency_ms"`
	ErrorMessage    string    `json:"error_message" yaml:"error_message"`
	DowntimeMinutes int       `json:"downtime_minutes" yaml:"downtime_minutes"`
	TotalDowntime   string    `json:"total_downtime" yaml:"total_downtime"`
	SentAt          time.Time `json:"sent_at" yaml:"sent_at"`
}
