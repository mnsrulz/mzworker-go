## Context

Every query — the `query` subcommand and all shortcut commands (`ohlc`, `options-stat`, `volatility`, `expected-move`, `bull-run-signal`) — converges on `query.ExecuteQuery`, which composes the final statement (CTE scaffold + command SQL + `LIMIT`) and calls `db.QueryContext`. The same path is used by AMQP `serve` mode because handlers resolve their executor through `handler.Init`.

Logging is currently done with the standard `log` package in 23 places across `main.go`, `cmd/serve.go`, `amqp/consumer.go`, `amqp/dispatch.go`, `handler/timing.go`, and `handler/validation.go`. There are no levels, so there is no knob to expose additional diagnostic detail such as the executed SQL. The project targets Go 1.27, so `log/slog` is available in the standard library.

## Goals / Non-Goals

**Goals:**
- Introduce leveled logging with a single configuration point.
- Emit the exact SQL sent to DuckDB at `Debug`.
- Raise verbosity via `-v` / `--verbose` for CLI commands and `LOG_LEVEL=debug` for the daemon.
- Keep stdout clean for `json`, `table`, and `raw` output.
- Preserve default (`Info`) visibility of existing lifecycle, timing, validation, and error records.

**Non-Goals:**
- Structured JSON logging, log sampling, or log-file configuration.
- Logging result sets, query parameters, or per-phase timings.
- Third-party logging dependencies.
- Changing AMQP payloads, reply envelopes, or output formats.

## Decisions

### D1: Use the standard library `log/slog`
Go 1.27 ships `log/slog`, which provides levels, a text handler, and structured attributes with no new dependency. A default `slog` logger writing to stderr is the single logging sink.

**Alternatives considered (rejected):** a hand-rolled level wrapper over `log` (reinvents slog and loses structured attributes); a third-party logger such as zap/zerolog (unnecessary dependency for this project size); keeping the boolean gate (no headroom for other verbosity).

### D2: Configure the level once in `RootCmd.PersistentPreRunE`
The pre-run configures `slog.SetDefault` with `slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: ...})`. Level resolution:
1. `-v` / `--verbose` forces `Debug` (explicit flag wins).
2. Otherwise `LOG_LEVEL` is parsed (`debug|info|warn|error`, case-insensitive).
3. Otherwise `Info`.
An unparsable `LOG_LEVEL` falls back to `Info` and logs a warning.

`PersistentPreRunE` applies to every subcommand and runs after flag parsing, so both CLI flags and environment are available.

### D3: Query SQL is a `Debug` record at the single choke point
In `query.ExecuteQuery`, immediately before `db.QueryContext`: `slog.Debug("executing query", "sql", sqlWithLimit)`. The logger's level is the gate — there is no separate boolean, setter, or global flag. `ExecuteQuery` is the one path used by every command and transport, and `sqlWithLimit` is byte-for-byte what DuckDB receives.

**Alternative considered (rejected):** logging in `InternalQueryExecutor.Execute` — captures the pre-`LIMIT` string and misses direct `ExecuteQuery` callers.

### D4: Persistent `-v` / `--verbose` flag on the root command
One declaration makes the flag available to every subcommand; the pre-run maps it to `Debug`. `-v` is unused today — Cobra's automatic version flag is long-only `--version`.

**Alternative considered (rejected):** per-command flags — repetitive across six commands with identical semantics.

### D5: `LOG_LEVEL` environment variable for the daemon
`serve` has no per-request CLI invocation in containers, so `LOG_LEVEL=debug` raises verbosity there. The variable is honored by the shared pre-run, so it works for CLI commands too.

### D6: Migrate all existing `log.*` calls to `slog`
Existing lines become `slog.Info` for lifecycle/timing/validation (preserving default visibility), `slog.Warn`/`slog.Error` for failure paths, and `log.Fatal`/`log.Fatalf` become `slog.Error` followed by `os.Exit(1)`. `amqp.SetLogger(log.Default())` stays because the RabbitMQ client requires a `*log.Logger`; its internal logs remain outside slog levels.

## Risks / Trade-offs

- [Log text format changes for every line] → Accepted; the `Pipeline:`/`AMQP:` message text is preserved so existing greps keep working.
- [RabbitMQ client logs are not level-controlled] → Accepted; they stay on the standard `log` package.
- [Fatal paths no longer use `log.Fatal`] → `slog.Error` + `os.Exit(1)` reproduces the same exit behavior.
- [Full CTE makes the debug SQL record large] → Debug-only, so production/default runs are unaffected.
- [Tests that captured `log` output break] → `handler/timing_test.go` is updated to capture a `slog` handler.

## Migration Plan

Additive: default `Info` output is preserved, and only the SQL record is new and hidden behind `Debug`. Rollback is reverting the logging migration and the debug record. No data or schema migration.

## Open Questions

- None.
