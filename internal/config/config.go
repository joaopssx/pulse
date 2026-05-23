package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services  []ServiceConfig `yaml:"services"`
	Alerts    AlertsConfig    `yaml:"alerts"`
	Dashboard DashboardConfig `yaml:"dashboard"`
	Storage   StorageConfig   `yaml:"storage"`
	Baseline  BaselineConfig  `yaml:"baseline"`
}

type ServiceConfig struct {
	Name           string        `yaml:"name"`
	URL            string        `yaml:"url"`
	Interval       time.Duration `yaml:"interval"`
	Timeout        time.Duration `yaml:"timeout"`
	ExpectedStatus int           `yaml:"expected_status"`
	Tags           []string      `yaml:"tags"`
	AlertChannels  []string      `yaml:"alert_channels"`
	Enabled        bool          `yaml:"enabled"`
}

type AlertsConfig struct {
	Bot BotConfig `yaml:"bot"`
}

type BotConfig struct {
	Endpoint string `yaml:"endpoint"`
	Secret   string `yaml:"secret"`
}

type DashboardConfig struct {
	Port    int  `yaml:"port"`
	Enabled bool `yaml:"enabled"`
}

type StorageConfig struct {
	Path string `yaml:"path"`
}

type BaselineConfig struct {
	WindowSize          int     `yaml:"window_size"`
	ThresholdMultiplier float64 `yaml:"threshold_multiplier"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
