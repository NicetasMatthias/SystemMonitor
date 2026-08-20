VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%d %H:%M:%S")

INFO_PKG := github.com/NicetasMatthias/SystemMonitor/internal/info

LDFLAGS := \
	-X $(INFO_PKG).Version=$(VERSION) \
	-X $(INFO_PKG).Commit=$(COMMIT) \
	-X $(INFO_PKG).Date=$(DATE)

ifeq ($(OS),Windows_NT)
    BINARY_NAME=system-monitor.exe
else
    BINARY_NAME=system-monitor
endif

BINARY_DIR=bin
MAIN_PACKAGE=./cmd/system-monitor

.PHONY: all build clean run run-prod deps help test test-race lint info

all: deps build

help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: deps
	@echo "Building..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Done! Binary: $(BINARY_DIR)/$(BINARY_NAME)"
	
run: deps
	go run -tags "dev" $(MAIN_PACKAGE)
	
run-prod: deps
	go run $(MAIN_PACKAGE)

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_DIR)
	go clean

test: deps
	go test ./...

test-race: deps
	go test -race ./...

lint: deps
	golangci-lint run

info:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"
