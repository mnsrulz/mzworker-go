APP_NAME := mzworker
BIN_DIR := bin

.PHONY: all gen build test lint clean cross

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
	rm -rf $(BIN_DIR)

# Cross-compilation targets
cross:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-musl-gcc go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-musl-gcc go build -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe .
