APP_NAME := mzworker
BIN_DIR := bin

.PHONY: all build test lint clean

all: build

build:
	go build -o $(BIN_DIR)/$(APP_NAME) .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf $(BIN_DIR) dist
