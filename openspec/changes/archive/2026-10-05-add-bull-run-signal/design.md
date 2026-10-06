## Context

The bull-run signal series SQL (weekly cheap-vol/positioning/wing checklist + trend gate, forwarded returns) currently lives as a loose file `~/mz-signal-series.sql` outside any repo. Consumers run it through the generic `dynamic-sql-query` request, which accepts arbitrary SQL and wraps it in the CTE scaffold (`expirations, T, T2, base, calc, ranked, dataset`) built by `query/internal.go`. The worker already has four domain commands (`ohlc`, `volatility`, `expected-move`, `options-stat`) that follow one pattern: cobra command → `registerMediatr` → `mediatr.Send` → handler builds SQL → `QueryExecutor` → `Output`. The same handlers are exposed over AMQP because `handler.RegisterStruct` registers the request factory (AMQP dispatch) and the mediatr handler in one call.

## Goals / Non-Goals

**Goals:**
- Make the bull-run series SQL a versioned, handler-owned constant in this repo
- Expose it as CLI `mzworker-go bull-run-signal` and AMQP `bull-run-signal-query`
- Parameterize the date range with the repo-wide `--lookback` convention
- Guarantee the same rows/columns consumers get today via `dynamic-sql-query`

**Non-Goals:**
- Changing the SQL's scoring/labeling semantics or column set
- Modifying the scaffold, AMQP envelope, or any existing request type
- Swapping mytradingview's app client over to the new request type (follow-up, separate repo)
- Redeploying `serve` (Docker/GHCR) — operational follow-up after merge

## Decisions

- **D1 — One registration serves both surfaces.** `RegisterStruct[BullRunSignalQuery, *QueryResponse]("bull-run-signal-query", ...)` in the handler's `init()` gives AMQP factory+dispatch (`amqp/dispatch.go`) and the mediatr handler the CLI uses. Alternative: separate CLI/AMQP wiring — rejected as duplication with no benefit.
- **D2 — SQL as an inline raw-string const** (`bullRunSeriesSQL`) in `handler/bull_run_signal_handler.go`, matching every existing handler (no `go:embed` anywhere in the repo). The const is byte-for-byte the file SQL except the final `WHERE wk >= DATE '2025-01-01'` becomes `WHERE wk >= (current_date - %d)`. Verified: no `%` or backtick characters, so `fmt.Sprintf` is safe.
- **D3 — `--lookback` / `-d`, default 730.** Flag name/shorthand match all four existing commands. Default deviates from their `30` because 30 days can never satisfy the signal's 17-week trend gate; 730 ≈ full ODATA history, matching today's `wk >= 2025-01-01` output. The filter stays in the final `WHERE`, so window functions (z-scores, SMAs, `LEAD`) still compute over the full history before trimming — identical semantics to the current file.
- **D4 — Reuse the existing reply envelope.** `toAmqpResponse` already handles `*QueryResponse` (columnar, column order preserved); no `amqp/` changes.
- **D5 — No `limit` field on the request.** Row count is bounded by `--lookback` (weeks, not rows); handler passes `99999` like `volatility`, capped at 10000 by the scaffold — mirrors existing commands.

## Risks / Trade-offs

- [Row order through the scaffold subquery wrapper] → The wrapper is `SELECT * FROM ( <sql> ) LIMITED_CTE LIMIT n`; order is not formally guaranteed. Mitigation: this exact path already delivered ordered rows via `dynamic-sql-query`; the parity verification task compares first/last week against the current path.
- [Short lookback yields an insufficient series (trend gate never turns ON)] → Accepted; consumers already enforce `MIN_BULL_RUN_WEEKS` client-side. Server returns whatever exists.
- [golangci-lint not installed on this machine] → `make lint` unavailable; use `go vet ./...` plus `go test ./...` as the gate, note in tasks.
- [Parallel definition drift if the stray `~/mz-signal-series.sql` stays] → Delete it after parity passes (task in this change).

## Migration Plan

Purely additive — new files only plus one `types.go` block. Existing request types untouched, so rollback is removing the new files. `serve` keeps working for current callers until redeployed; the new request type simply returns "unknown request type" on old binaries.

## Open Questions

None.
