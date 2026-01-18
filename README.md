# Hydra (Go)

A configuration-driven HTML processing engine inspired by the Python [Hydra](https://hydra.cc) framework.

## Overview

This project implements a Go-based version of the core Hydra philosophy: **Composition over Configuration**. It allows users to define complex web scraping and data "elevation" tasks through external YAML configurations.

### Key Features

- **Hierarchical Configuration**: Manage complex settings through structured YAML.
- **Data Elevation**: Automatically "elevate" raw HTML responses into structured insights.
- **Modular Architecture**: Clean separation between fetching, evaluation, and configuration.

## Getting Started

### Prerequisites

- Go 1.22 or higher

### Installation

1. Clone the repository to your GoLandProjects directory.
2. Initialize dependencies:
   ```bash
   go mod tidy
   ```

### Running

To run the default configuration:
```bash
go run cmd/hydra/main.go configs/config.yaml
```

### Building

```bash
make build
```

## Configuration

The `configs/config.yaml` controls the behavior of the program:

```yaml
fetcher:
  url: "https://example.com"
  timeout_seconds: 10

evaluator:
  target_selector: "title"
  logic: "extract_text"
```

In this model, **Elevation** refers to the process of promoting low-level HTTP responses into high-level business entities or data points.

## Project Structure

- `cmd/hydra/`: Main entry point.
- `internal/config/`: Configuration loading and validation.
- `internal/fetcher/`: HTTP client logic.
- `internal/evaluator/`: HTML parsing and data elevation logic.
- `configs/`: YAML configuration files.
