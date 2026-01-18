.PHONY: build run test clean

BINARY_NAME=hydra

build:
	go build -o bin/$(BINARY_NAME) cmd/hydra/main.go

run:
	go run cmd/hydra/main.go configs/config.yaml

test:
	go test ./...

clean:
	rm -rf bin/
	go clean
