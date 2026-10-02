APP_NAME := mzworker-go
BIN_DIR := bin

VERSION ?= $(shell git describe --tags --long --match "v[0-9]*.[0-9]*.[0-9]*" --always --dirty 2>/dev/null || echo dev)
GIT_SHA ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.Version=$(VERSION) -X main.GitSHA=$(GIT_SHA) -X main.BuildTime=$(BUILD_TIME)

.PHONY: all build test lint clean

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf $(BIN_DIR) dist
