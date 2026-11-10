## Why

There is no way to see the SQL that is actually sent to DuckDB, and logging has no verbosity levels — it is ad-hoc `log.Printf` output that is either always on or absent. Shortcut commands hide their query behind a large CTE scaffold, and the existing `Pipeline:` lines only show request type and duration. Introducing level-based logging and emitting the executed SQL at debug level makes debugging and verification straightforward without adding a one-off flag.

## What Changes

- Adopt the standard library `log/slog` as the project's logging layer, writing leveled records to stderr with a default minimum level of `Info`.
- Add a persistent `-v` / `--verbose` flag to the root command that raises the minimum level to `Debug`.
- Support a `LOG_LEVEL` environment variable (`debug`, `info`, `warn`, `error`, case-insensitive) for `serve` and container deployments; an invalid value falls back to `Info` with a warning.
- Log the exact composed SQL — CTE scaffold, command SQL, and `LIMIT` — at `Debug`, immediately before execution, so it appears only when verbosity is raised.
- Migrate the existing `log.Printf` / `log.Fatal` call sites to `slog`, keeping their default visibility at `Info` (failure paths at `Error`/`Warn`).
- Keep all log output on stderr so stdout (`json`, `table`, `raw`) stays machine-readable.

## Capabilities

### New Capabilities
- `verbose-query-logging`: Level-based logging with the full SQL statement emitted at `Debug`, controlled by a CLI flag and an environment variable.

### Modified Capabilities

(None — no existing requirement changes.)

## Impact

- `cmd/root.go` — persistent `--verbose` flag and slog configuration in `PersistentPreRunE`, including `LOG_LEVEL` parsing.
- `cmd/serve.go` — logging migrated to `slog`; `LOG_LEVEL` already applied by the root pre-run.
- `query/executor.go` — `slog.Debug("executing query", "sql", ...)` at the single execution choke point.
- `handler/timing.go`, `handler/validation.go`, `amqp/consumer.go`, `amqp/dispatch.go`, `main.go` — `log.*` migrated to `slog`.
- `handler/timing_test.go` — log capture switched to a `slog` handler; `query/executor_test.go` — new debug-level assertions.
- `README.md` — usage, verbosity, and `LOG_LEVEL` documentation.
- No changes to AMQP request/reply payloads or output formats.
- Backward compatible in visibility: default `Info` output still shows lifecycle, timing, and error records; only the new SQL line requires `-v` / `LOG_LEVEL=debug`. Log line formatting changes to slog's text format.
