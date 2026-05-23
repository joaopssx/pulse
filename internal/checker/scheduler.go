package checker

import (
	"context"
	"sync"
	"time"

	"github.com/pulse/pulse/internal/alert"
	"github.com/pulse/pulse/internal/models"
)

type Store interface {
	SaveCheckResult(ctx context.Context, result Result) error
}

type AnomalyDetector interface {
	Analyze(ctx context.Context, result Result)
}

type Scheduler struct {
	services   []models.Service
	store      Store
	dispatcher alert.Dispatcher
	detector   AnomalyDetector
	wg         sync.WaitGroup
	cancel     context.CancelFunc
}

func NewScheduler(services []models.Service, store Store, dispatcher alert.Dispatcher, detector AnomalyDetector) *Scheduler {
	return &Scheduler{
		services:   services,
		store:      store,
		dispatcher: dispatcher,
		detector:   detector,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	for _, service := range s.services {
		if !service.Enabled {
			continue
		}
		s.wg.Add(1)
		go s.runService(ctx, service)
	}
}

func (s *Scheduler) runService(ctx context.Context, service models.Service) {
	defer s.wg.Done()

	chk := NewChecker(service)
	ticker := time.NewTicker(service.Interval)
	defer ticker.Stop()

	s.runCheck(ctx, chk)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runCheck(ctx, chk)
		}
	}
}

func (s *Scheduler) runCheck(ctx context.Context, chk *Checker) {
	res := chk.Check(ctx)
	_ = s.store.SaveCheckResult(ctx, res)
	if s.detector != nil {
		s.detector.Analyze(ctx, res)
	}
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}
