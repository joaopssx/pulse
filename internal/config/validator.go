package config

import (
	"fmt"
	"net/url"
	"time"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func Validate(cfg *Config) []ValidationError {
	var errors []ValidationError

	names := make(map[string]bool, len(cfg.Services))
	usesBot := false

	for i, svc := range cfg.Services {
		prefix := fmt.Sprintf("services[%d]", i)

		if svc.Name == "" {
			errors = append(errors, ValidationError{Field: prefix + ".name", Message: "must not be empty"})
		} else if names[svc.Name] {
			errors = append(errors, ValidationError{Field: prefix + ".name", Message: "must be unique"})
		} else {
			names[svc.Name] = true
		}

		if svc.URL == "" {
			errors = append(errors, ValidationError{Field: prefix + ".url", Message: "must not be empty"})
		} else {
			u, err := url.ParseRequestURI(svc.URL)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				errors = append(errors, ValidationError{Field: prefix + ".url", Message: "must be a valid HTTP/HTTPS URL"})
			}
		}

		if svc.Interval == 0 {
			errors = append(errors, ValidationError{Field: prefix + ".interval", Message: "must not be empty"})
		} else if svc.Interval < 5*time.Second {
			errors = append(errors, ValidationError{Field: prefix + ".interval", Message: "must be >= 5s"})
		}

		if svc.Timeout == 0 {
			errors = append(errors, ValidationError{Field: prefix + ".timeout", Message: "must not be empty"})
		} else if svc.Timeout >= svc.Interval {
			errors = append(errors, ValidationError{Field: prefix + ".timeout", Message: "must be < interval"})
		}

		for _, ch := range svc.AlertChannels {
			if ch != "bot" {
				errors = append(errors, ValidationError{Field: prefix + ".alert_channels", Message: "must only contain known values: bot"})
			} else {
				usesBot = true
			}
		}
	}

	if usesBot && cfg.Alerts.Bot.Endpoint == "" {
		errors = append(errors, ValidationError{Field: "alerts.bot.endpoint", Message: "must not be empty when 'bot' channel is used"})
	}

	if cfg.Dashboard.Port < 1024 || cfg.Dashboard.Port > 65535 {
		errors = append(errors, ValidationError{Field: "dashboard.port", Message: "must be between 1024 and 65535"})
	}

	if cfg.Storage.Path == "" {
		errors = append(errors, ValidationError{Field: "storage.path", Message: "must not be empty"})
	}

	if cfg.Baseline.WindowSize <= 0 {
		errors = append(errors, ValidationError{Field: "baseline.window_size", Message: "must be > 0"})
	}

	if cfg.Baseline.ThresholdMultiplier <= 0 {
		errors = append(errors, ValidationError{Field: "baseline.threshold_multiplier", Message: "must be > 0"})
	}

	return errors
}
