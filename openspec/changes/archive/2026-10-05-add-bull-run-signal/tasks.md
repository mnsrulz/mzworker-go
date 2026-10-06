## 1. Handler

- [x] 1.1 Add `BullRunSignalQuery{Symbol, LookbackDays}` to `handler/types.go` with zog schema and `Validate()` (symbol required, lookbackDays required + positive)
- [x] 1.2 Create `handler/bull_run_signal_handler.go`: `bullRunSeriesSQL` const (file SQL verbatim, `WHERE wk >= (current_date - %d)`), `BullRunSignalHandler`, `init()` registering `"bull-run-signal-query"` via `RegisterStruct`, `Handle` → `fmt.Sprintf` + `executor(ctx, symbol, sql, 99999)`

## 2. CLI

- [x] 2.1 Create `cmd/bull_run_signal.go`: cobra `bull-run-signal` with `-s` (required), `--data-dir` (required), `-d/--lookback` default 730, `-f/--format` default json → `registerMediatr` → `mediatr.Send` → `Output`

## 3. Tests

- [x] 3.1 `handler/bull_run_signal_handler_test.go`: validation valid/invalid (missing symbol, non-positive lookback), `GetRequestFactory("bull-run-signal-query")` registered, dispatch through pipeline rejects invalid request
- [x] 3.2 SQL sanity test: const contains `dataset` and `rule_pass`, uses `current_date -` window, contains no hardcoded `2025-01-01`

## 4. Verification

- [x] 4.1 `go vet ./...` and `go test ./...` pass (note: `golangci-lint` not installed, `make lint` unavailable)
- [x] 4.2 `go build` produces `bin/mzworker-go`; `bull-run-signal --help` shows the new command with its flags
- [x] 4.3 IBIT parity: `bull-run-signal -s IBIT -d 730 --data-dir /Users/mz/ODATA` vs the same SQL through `query`/`dynamic-sql-query` — same row count, same first/last `wk`, spot-check `rule_pass` and scores
- [x] 4.4 Delete stray `~/mz-signal-series.sql` after parity passes
