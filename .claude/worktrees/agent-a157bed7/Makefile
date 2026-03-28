.PHONY: build test test-verbose lint clean

BINARY := reqflow
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/reqflow

test:
	go test ./...

test-verbose:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

lint:
	golangci-lint run

clean:
	rm -rf $(BUILD_DIR) coverage.txt coverage.html
