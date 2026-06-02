package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Upstreams []UpstreamConfig `yaml:"upstreams"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type UpstreamConfig struct {
	Name        string        `yaml:"name"`
	Strategy    string        `yaml:"strategy"`
	Targets     []TargetConfig `yaml:"targets"`
	HealthCheck HealthConfig  `yaml:"health_check"`
	Retries     int           `yaml:"retries"`
}

type TargetConfig struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type HealthConfig struct {
	Path     string        `yaml:"path"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	for i := range cfg.Upstreams {
		if cfg.Upstreams[i].Retries == 0 {
			cfg.Upstreams[i].Retries = 2
		}
		if cfg.Upstreams[i].Strategy == "" {
			cfg.Upstreams[i].Strategy = "round-robin"
		}
		if cfg.Upstreams[i].HealthCheck.Interval == 0 {
			cfg.Upstreams[i].HealthCheck.Interval = 10 * time.Second
		}
		if cfg.Upstreams[i].HealthCheck.Timeout == 0 {
			cfg.Upstreams[i].HealthCheck.Timeout = 3 * time.Second
		}
		for j := range cfg.Upstreams[i].Targets {
			if cfg.Upstreams[i].Targets[j].Weight == 0 {
				cfg.Upstreams[i].Targets[j].Weight = 1
			}
		}
	}

	return &cfg, nil
}
