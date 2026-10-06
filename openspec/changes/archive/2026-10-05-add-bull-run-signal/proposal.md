## Why

The bull-run entry signal (weekly cheap-vol / positioning / wing checklist with an uptrend gate) exists only as a loose SQL file (`~/mz-signal-series.sql`) that consumers must submit through the generic `dynamic-sql-query` request. The SQL is not versioned with the worker, the query name means nothing to callers, and every consumer has to know where the file lives. Making it a first-class command puts the SQL under repo control and gives both CLI and AMQP callers a named, validated interface.

## What Changes

- New CLI command `mzworker-go bull-run-signal` with `-s/--symbol`, `--data-dir`, `-d/--lookback` (days, default 730), `-f/--format`
- New AMQP request type `bull-run-signal-query` accepting `{symbol, lookbackDays}` and replying through the existing columnar reply envelope
- The series SQL moves into the worker as a handler-owned constant, with its hardcoded start date replaced by a `--lookback` window (`WHERE wk >= (current_date - N)`), matching how every other command parameterizes its date range
- No transport, envelope, or scaffold changes; response shape is the standard `QueryResponse` columnar value

## Capabilities

### New Capabilities

- `bull-run-signal`: the bull-run signal weekly series — SQL ownership, CLI surface, AMQP request type, lookback windowing, and input validation

### Modified Capabilities

(none — existing `dynamic-sql-query`, envelope, and timing behavior are unchanged)

## Impact

- `handler/types.go` — new `BullRunSignalQuery` request + zog schema
- `handler/bull_run_signal_handler.go` (new) — SQL constant, handler, `RegisterStruct` init
- `cmd/bull_run_signal.go` (new) — cobra command
- `handler/bull_run_signal_handler_test.go` (new) — validation, registration, SQL sanity tests
- No changes to `amqp/`, `query/` scaffold, or other commands; consumers of `dynamic-sql-query` keep working
- Follow-up (out of scope): mytradingview app swaps `runDynamicQuery` for `submitQuery('bull-run-signal-query', ...)`; Docker/GHCR redeploy of `serve`
