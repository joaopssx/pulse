package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pulse/pulse/internal/checker"
	"github.com/pulse/pulse/internal/models"
)

type Store interface {
	GetActiveIncidents(ctx context.Context) ([]models.Incident, error)
	GetIncidentsByService(ctx context.Context, serviceID string, limit int) ([]models.Incident, error)
	GetRecentResults(ctx context.Context, serviceID string, limit int) ([]checker.Result, error)
	GetUptimePercent(ctx context.Context, serviceID string, since time.Time) (float64, error)
}

type handler struct {
	store    Store
	services []models.Service
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) getServices(w http.ResponseWriter, r *http.Request) {
	type serviceRes struct {
		models.Service
		Status  string  `json:"status"`
		Uptime  float64 `json:"uptime_24h"`
	}

	res := make([]serviceRes, 0, len(h.services))
	for _, svc := range h.services {
		uptime, err := h.store.GetUptimePercent(r.Context(), svc.ID, time.Now().Add(-24*time.Hour))
		if err != nil {
			uptime = 0
		}

		status := string(models.ServiceStatusUnknown)
		results, err := h.store.GetRecentResults(r.Context(), svc.ID, 1)
		if err == nil && len(results) > 0 {
			if results[0].Success {
				status = string(models.ServiceStatusHealthy)
			} else {
				status = string(models.ServiceStatusDown)
			}
		}

		res = append(res, serviceRes{
			Service: svc,
			Status:  status,
			Uptime:  uptime,
		})
	}

	respondJSON(w, http.StatusOK, res)
}

func (h *handler) getServiceIncidents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	incidents, err := h.store.GetIncidentsByService(r.Context(), id, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get incidents")
		return
	}
	if incidents == nil {
		incidents = []models.Incident{}
	}
	respondJSON(w, http.StatusOK, incidents)
}

func (h *handler) getServiceResults(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	results, err := h.store.GetRecentResults(r.Context(), id, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get results")
		return
	}
	if results == nil {
		results = []checker.Result{}
	}
	respondJSON(w, http.StatusOK, results)
}

func (h *handler) getActiveIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.store.GetActiveIncidents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get active incidents")
		return
	}
	if incidents == nil {
		incidents = []models.Incident{}
	}
	respondJSON(w, http.StatusOK, incidents)
}
