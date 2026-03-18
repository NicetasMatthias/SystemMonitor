BINARY_NAME=system-monitor
BINARY_DIR=bin
MAIN_PACKAGE=./cmd/system-monitor

.PHONY: all build clean test run deps help

all: deps build

help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: deps
	@echo "Building..."
	go build -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Done! Binary: $(BINARY_DIR)/$(BINARY_NAME)"

run: build
	@echo "Running..."
	./$(BINARY_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test -v -race ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_DIR)
	go clean