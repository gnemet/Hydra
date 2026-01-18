.PHONY: build run test clean build-all run-test-env

HYDRA_BIN=bin/hydra
SERVER_BIN=bin/testserver

build:
	mkdir -p bin
	go build -o $(HYDRA_BIN) cmd/hydra/main.go

build-server:
	mkdir -p bin
	go build -o $(SERVER_BIN) cmd/testserver/main.go

build-all: build build-server

run: build
	./$(HYDRA_BIN) configs/config.yaml

run-test-config: build
	./$(HYDRA_BIN) configs/test_config.yaml

run-server: build-server
	./$(SERVER_BIN)

test:
	go test ./...

clean:
	rm -rf bin/
	go clean
