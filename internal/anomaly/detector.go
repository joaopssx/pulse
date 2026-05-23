package anomaly

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/pulse/pulse/internal/alert"
	"github.com/pulse/pulse/internal/checker"
	"github.com/pulse/pulse/internal/models"
)

type Store interface {
	GetIncidentsByService(ctx context.Context, serviceID string, limit int) ([]models.Incident, error)
	SaveIncident(ctx context.Context, incident models.Incident) error
	ResolveIncident(ctx context.Context, incidentID string, resolvedAt time.Time) error
}

type Detector struct {
	baseline            *Baseline
	thresholdMultiplier float64
	store               Store
	dispatcher          alert.Dispatcher
	slowCooldowns       sync.Map
}

func NewDetector(baseline *Baseline, thresholdMultiplier float64, store Store, dispatcher alert.Dispatcher) *Detector {
	return &Detector{
		baseline:            baseline,
		thresholdMultiplier: thresholdMultiplier,
		store:               store,
		dispatcher:          dispatcher,
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (d *Detector) Analyze(ctx context.Context, result checker.Result) {
	incidents, err := d.store.GetIncidentsByService(ctx, result.ServiceID, 1)
	if err != nil {
		return
	}

	var activeIncident *models.Incident
	if len(incidents) > 0 && !incidents[0].Resolved {
		activeIncident = &incidents[0]
	}

	if !result.Success {
		if activeIncident == nil {
			inc := models.Incident{
				ID:          generateID(),
				ServiceID:   result.ServiceID,
				ServiceName: result.ServiceName,
				StartedAt:   result.CheckedAt,
				Cause:       result.Error,
				StatusCode:  result.StatusCode,
				LatencyMs:   result.LatencyMs,
				Resolved:    false,
			}
			_ = d.store.SaveIncident(ctx, inc)

			al := models.Alert{
				IncidentID:   inc.ID,
				ServiceName:  result.ServiceName,
				Status:       string(models.AlertTypeDown),
				URL:          result.URL,
				LatencyMs:    result.LatencyMs,
				ErrorMessage: result.Error,
				SentAt:       time.Now(),
			}
			_ = d.dispatcher.Send(ctx, al)
		}
		return
	}

	if activeIncident != nil {
		_ = d.store.ResolveIncident(ctx, activeIncident.ID, result.CheckedAt)
		downtime := result.CheckedAt.Sub(activeIncident.StartedAt)

		al := models.Alert{
			IncidentID:      activeIncident.ID,
			ServiceName:     result.ServiceName,
			Status:          string(models.AlertTypeRecovered),
			URL:             result.URL,
			LatencyMs:       result.LatencyMs,
			DowntimeMinutes: int(downtime.Minutes()),
			TotalDowntime:   fmt.Sprintf("%d minutes", int(downtime.Minutes())),
			SentAt:          time.Now(),
		}
		_ = d.dispatcher.Send(ctx, al)
		return
	}

	if d.baseline.SampleCount(result.ServiceID) >= 10 {
		mean := d.baseline.Mean(result.ServiceID)
		stddev := d.baseline.StdDev(result.ServiceID)

		if float64(result.LatencyMs) > mean+(stddev*d.thresholdMultiplier) {
			lastAlertVal, ok := d.slowCooldowns.Load(result.ServiceID)
			canAlert := true
			if ok {
				lastAlertTime := lastAlertVal.(time.Time)
				if time.Since(lastAlertTime) < 5*time.Minute {
					canAlert = false
				}
			}

			if canAlert {
				d.slowCooldowns.Store(result.ServiceID, time.Now())
				al := models.Alert{
					ServiceName:     result.ServiceName,
					Status:          string(models.AlertTypeSlow),
					URL:             result.URL,
					LatencyMs:       result.LatencyMs,
					NormalLatencyMs: int64(mean),
					SentAt:          time.Now(),
				}
				_ = d.dispatcher.Send(ctx, al)
			}
		}
	}

	d.baseline.Add(result.ServiceID, result.LatencyMs)
}
