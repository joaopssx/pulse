package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pulse/pulse/internal/alert"
	"github.com/pulse/pulse/internal/anomaly"
	"github.com/pulse/pulse/internal/checker"
	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/dashboard"
	"github.com/pulse/pulse/internal/models"
	"github.com/pulse/pulse/internal/storage"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	configPath := flag.String("config", "config.yaml", "")
	port := flag.Int("port", 0, "")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	valErrs := config.Validate(cfg)
	if len(valErrs) > 0 {
		var msgs []string
		for _, e := range valErrs {
			msgs = append(msgs, e.Error())
		}
		slog.Error("invalid configuration", "errors", strings.Join(msgs, ", "))
		os.Exit(1)
	}

	store, err := storage.New(cfg.Storage.Path)
	if err != nil {
		slog.Error("failed to init storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	baseline := anomaly.NewBaseline(cfg.Baseline.WindowSize)

	botDispatcher := alert.NewBotDispatcher(cfg.Alerts.Bot.Endpoint, cfg.Alerts.Bot.Secret, 10*time.Second)
	multiDispatcher := alert.NewMultiDispatcher(botDispatcher)

	detector := anomaly.NewDetector(baseline, cfg.Baseline.ThresholdMultiplier, store, multiDispatcher)

	services := make([]models.Service, 0, len(cfg.Services))
	for _, sc := range cfg.Services {
		services = append(services, models.Service{
			ID:             sc.Name,
			Name:           sc.Name,
			URL:            sc.URL,
			Interval:       sc.Interval,
			Timeout:        sc.Timeout,
			ExpectedStatus: sc.ExpectedStatus,
			Tags:           sc.Tags,
			AlertChannels:  sc.AlertChannels,
			Enabled:        sc.Enabled,
			CreatedAt:      time.Now(),
		})
	}

	dashPort := cfg.Dashboard.Port
	if *port != 0 {
		dashPort = *port
	}

	scheduler := checker.NewScheduler(services, store, multiDispatcher, detector)
	dashServer := dashboard.New(dashPort, store, services)

	slog.Info("starting pulse",
		slog.Int("services_count", len(services)),
		slog.Int("dashboard_port", dashPort),
		slog.String("bot_endpoint", cfg.Alerts.Bot.Endpoint),
	)

	ctx, cancel := context.WithCancel(context.Background())

	go scheduler.Start(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := dashServer.Start(ctx); err != nil {
			slog.Error("dashboard server error", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	cancel()

	done := make(chan struct{})
	go func() {
		scheduler.Stop()
		wg.Wait()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		slog.Error("graceful shutdown timed out")
		os.Exit(1)
	case <-done:
	}
}
