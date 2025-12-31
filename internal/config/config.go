package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config mirrors the existing appsettings.json fields expected by the exporter.
type Config struct {
	BBoxAPIURL                 string `json:"BBoxAPIURL"`
	BBoxPassword               string `json:"BBoxPassword"`
	BBoxAPIRefreshTime         int    `json:"BBoxAPIRefreshTime"`
	MetricsServerListeningPort int    `json:"MetricsServerListeningPort"`
}

// Load reads configuration from disk and applies minimal validation/defaults.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if pwd, ok := os.LookupEnv("BBOX_PASSWORD"); ok {
		cfg.BBoxPassword = pwd
	}

	if cfg.BBoxAPIURL == "" {
		return Config{}, fmt.Errorf("BBoxAPIURL is required")
	}
	if cfg.BBoxPassword == "" {
		return Config{}, fmt.Errorf("BBoxPassword is required")
	}
	if cfg.BBoxAPIRefreshTime <= 0 {
		cfg.BBoxAPIRefreshTime = int((60 * time.Second).Seconds())
	}
	if cfg.MetricsServerListeningPort == 0 {
		cfg.MetricsServerListeningPort = 9100
	}

	return cfg, nil
}
