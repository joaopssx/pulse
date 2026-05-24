package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pulse/pulse/internal/models"
)

type Server struct {
	port   int
	server *http.Server
}

func New(port int, store Store, services []models.Service) *Server {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, req)
		})
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, req)
		})
	})

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method not allowed"}`))
	})

	h := &handler{
		store:    store,
		services: services,
	}

	r.Get("/health", h.health)
	r.Get("/api/services", h.getServices)
	r.Get("/api/services/{id}/incidents", h.getServiceIncidents)
	r.Get("/api/services/{id}/results", h.getServiceResults)
	r.Get("/api/incidents/active", h.getActiveIncidents)

	addr := fmt.Sprintf(":%d", port)
	return &Server{
		port: port,
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
