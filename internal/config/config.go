package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Fetcher struct {
		URL     string `yaml:"url"`
		Timeout int    `yaml:"timeout_seconds"`
	} `yaml:"fetcher"`
	Evaluator struct {
		TargetSelector string `yaml:"target_selector"`
		Logic          string `yaml:"logic"`
	} `yaml:"evaluator"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
