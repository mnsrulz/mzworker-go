APP_NAME := mzworker
BIN_DIR := bin

.PHONY: all gen build test lint clean

all: build

gen:
	protoc --go_out=. --go_opt=module=github.com/mnsrulz/mzworker-go \
		--go-grpc_out=. --go-grpc_opt=module=github.com/mnsrulz/mzworker-go \
		proto/options.proto

build:
	go build -o $(BIN_DIR)/$(APP_NAME) .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf $(BIN_DIR) dist
