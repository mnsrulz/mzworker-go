## Why

The gRPC server transport is unnecessary complexity for this project's actual use case. The binary needs to be callable from opencode MCP/skills via shell invocation — a CLI that accepts raw SQL and returns results is the right interface. The gRPC server, proto definitions, and generated stubs add build tooling overhead (protoc, code generators) and dependency weight for a transport that isn't used.

## What Changes

- **BREAKING**: Remove gRPC server entirely (`server.go`, `proto/options.proto`, `gen/options/`)
- **BREAKING**: Remove `google.golang.org/grpc` and `google.golang.org/protobuf` dependencies
- Add CLI mode with urfave/cli: default action for one-shot query execution, `serve` subcommand for AMQP daemon
- Add `--format` flag: `json` (default, flat array of objects), `table` (ASCII table), `raw` (nested columns/rows for backward compat)
- Add `--limit` flag to override row limit from CLI
- Support SQL from command line argument or piped stdin
- Update `Makefile` to remove protobuf codegen targets
- Update `README.md` for CLI-first usage

## Capabilities

### New Capabilities
- `cli-query`: CLI interface for one-shot SQL query execution with formatted output (JSON, table, raw)

### Modified Capabilities

## Impact

- **Deleted files**: `server.go`, `gen/options/options.pb.go`, `gen/options/options_grpc.pb.go`, `proto/options.proto`
- **New file**: `cli.go`
- **Modified files**: `main.go` (subcommand routing), `Makefile` (remove gen target), `README.md` (rewrite)
- **Dependencies removed**: `google.golang.org/grpc`, `google.golang.org/protobuf`, `google.golang.org/genproto` (indirect)
- **Dependencies added**: `github.com/urfave/cli/v3`
- **Unchanged**: `amqp.go`, `handler/` package, `query.go`, all tests
