package main

import (
	"log"
	"os"

	"hydra/internal/config"
	"hydra/internal/evaluator"
	"hydra/internal/fetcher"
)

func main() {
	configPath := "configs/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	log.Printf("Starting Hydra with config: %s", configPath)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
	}

	f := fetcher.NewFetcher(cfg.Fetcher.Timeout)
	content, err := f.Fetch(cfg.Fetcher.URL)
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}

	e := evaluator.NewEvaluator(cfg.Evaluator.TargetSelector)
	result, err := e.Elevate(content)
	if err != nil {
		log.Fatalf("Evaluation error: %v", err)
	}

	log.Printf("Successfully elevated response!")
	log.Printf("Extracted [%s]: %s", cfg.Evaluator.TargetSelector, result)
}
