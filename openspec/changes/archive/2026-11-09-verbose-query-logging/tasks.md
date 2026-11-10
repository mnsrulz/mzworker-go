## 1. Logging foundation

- [x] 1.1 Add persistent `-v` / `--verbose` flag (default `false`) to `RootCmd` in `cmd/root.go`
- [x] 1.2 Add a `RootCmd.PersistentPreRunE` that configures `slog.SetDefault` with a stderr text handler and resolves the level: `-v` → Debug, else `LOG_LEVEL`, else Info
- [x] 1.3 Parse `LOG_LEVEL` (`debug|info|warn|error`, case-insensitive); fall back to Info and emit a warning on invalid values

## 2. Query logging

- [x] 2.1 In `query.ExecuteQuery`, log `slog.Debug("executing query", "sql", sqlWithLimit)` immediately before `db.QueryContext`
- [x] 2.2 Remove the temporary `atomic.Bool` / `SetLogQueries` idea; the level is the only gate

## 3. Migrate existing logging

- [x] 3.1 Migrate `handler/validation.go` and `handler/timing.go` to `slog.Info` (timing failures to `slog.Error`)
- [x] 3.2 Migrate `amqp/dispatch.go` failures to `slog.Error`/`slog.Warn` and success to `slog.Info`
- [x] 3.3 Migrate `amqp/consumer.go`: start/timeout to `slog.Info`/`slog.Warn`, terminal `log.Fatalf` to `slog.Error` + `os.Exit(1)`
- [x] 3.4 Migrate `cmd/serve.go` build info and lifecycle messages to `slog.Info`
- [x] 3.5 Migrate `main.go` `log.Fatal` to `slog.Error` + `os.Exit(1)`; keep `amqp.SetLogger(log.Default())`

## 4. Tests

- [x] 4.1 Update `handler/timing_test.go` to capture `slog` output instead of `log.SetOutput`
- [x] 4.2 Add `query/executor_test.go` test: at Debug level, an in-memory `SELECT 1 AS x` logs `msg="executing query"` with `sql="SELECT 1 AS x LIMIT 1000"`
- [x] 4.3 Add test: at Info level, no executed-SQL record is emitted
- [x] 4.4 Add a test for `LOG_LEVEL` parsing, including the invalid-value fallback

## 5. Documentation and verification

- [x] 5.1 Document `-v` / `--verbose`, `LOG_LEVEL`, and debug-level SQL logging in `README.md`
- [x] 5.2 Run `make test` and `make lint`
- [x] 5.3 Manually verify `-v` keeps `--format json` stdout valid, `2>/dev/null` suppresses logs, and `LOG_LEVEL=debug serve` enables SQL logging
