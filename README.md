# mzworker-go

gRPC server for querying DuckDB with SQL.

## Prerequisites

- Go 1.21+
- protoc (Protocol Buffers compiler)
- protoc-gen-go and protoc-gen-go-grpc

## Setup

```bash
# Install protobuf tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf code
make gen

# Build
make build

# Run
./bin/mzworker
```

## Environment Variables

- `LISTEN_ADDR` - Listen address (default: `0.0.0.0:50051`)
- `DATA_DIR` - Data directory path (default: `data/options_data`)

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Cross-compile for all platforms
make cross
```

## Docker

```bash
docker build -t mzworker .
docker run -p 50051:50051 -v /path/to/data:/data mzworker
```
