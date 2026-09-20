## 1. Dependencies

- [x] 1.1 Add `github.com/urfave/cli/v3` via `go get`
- [x] 1.2 Remove `google.golang.org/grpc` via `go mod tidy` (after deleting gRPC files)

## 2. Delete gRPC files

- [x] 2.1 Delete `server.go`
- [x] 2.2 Delete `gen/options/options.pb.go`
- [x] 2.3 Delete `gen/options/options_grpc.pb.go`
- [x] 2.4 Delete `proto/options.proto`

## 3. CLI Implementation

- [x] 3.1 Create `cli.go` with urfave/cli app definition
- [x] 3.2 Implement default action: read SQL from arg or stdin, execute via mediatr, format output
- [x] 3.3 Implement `serve` subcommand: start AMQP consumer with existing logic
- [x] 3.4 Implement `ohlc` subcommand: query OHLC data for a symbol
- [x] 3.5 Implement `formatJSON` — flat array of `map[string]interface{}` objects
- [x] 3.6 Implement `formatTable` — ASCII table using `text/tabwriter`
- [x] 3.7 Implement `formatRaw` — nested `{columns, rows}` for backward compat
- [x] 3.8 Implement stdin detection — show usage when no args + stdin is terminal

## 4. Handler Types

- [x] 4.1 Add `Symbol` field to `DynamicSQLQuery` struct
- [x] 4.2 Add `Symbol` to `DynamicSQLQuerySchema` validation
- [x] 4.3 Create `OHLCQuery` struct with `Symbol`, `From`, `To`, `Limit` fields
- [x] 4.4 Create `OHLCQuerySchema` validation schema
- [x] 4.5 Create `OHLCHandler` with `buildOHLCQuery` function

## 5. Registry and Dispatch

- [x] 5.1 Add `"ohlc-query"` to AMQP registry
- [x] 5.2 Add `OHLCQuery` case to `dispatchMediatr` in amqp.go
- [x] 5.3 Register `OHLCHandler` in `registerMediatr` in cli.go

## 6. Main entrypoint

- [x] 6.1 Rewrite `main.go` to wire urfave/cli app
- [x] 6.2 Remove gRPC server setup (`grpc.NewServer`, `pb.Register`, `reflection.Register`, `s.Serve`)
- [x] 6.3 Remove gRPC imports (`google.golang.org/grpc`, `google.golang.org/grpc/reflection`, `gen/options`)

## 7. Makefile and README

- [x] 7.1 Remove `gen` and `install-tools` targets from Makefile
- [x] 7.2 Rewrite README.md for CLI usage

## 8. Verification

- [x] 8.1 Run `go mod tidy` to clean up dependencies
- [x] 8.2 Run `go build ./...` to verify compilation
- [x] 8.3 Run `go vet ./...` and `go test ./...` to verify no regressions
- [x] 8.4 Run `golangci-lint run ./...` to verify lint passes
