## ADDED Requirements

### Requirement: CLI bull-run-signal command
The system SHALL provide a CLI command `mzworker-go bull-run-signal` that accepts a required `--symbol`/`-s`, a required `--data-dir`, a `--lookback`/`-d` day count (default 730), and a `--format`/`-f` of `json`, `table`, or `raw`, and prints the weekly bull-run signal series for that symbol.

#### Scenario: Successful run with defaults
- **WHEN** the user runs `mzworker-go bull-run-signal -s IBIT --data-dir /path/to/ODATA`
- **THEN** the command executes the series query against DuckDB using a 730-day lookback
- **AND** prints a JSON array of weekly rows ordered oldest to newest, each containing `wk`, `px`, `score`, `b_cheap`, `b_drought`, `b_dryup`, `b_wing`, `trend_up`, `rule_pass`, `fwd4w_pct`, and `fwd8w_pct`

#### Scenario: Table output format
- **WHEN** the user runs `mzworker-go bull-run-signal -s IBIT --data-dir /path/to/ODATA -f table`
- **THEN** the same rows are printed as a table

#### Scenario: Missing required flag
- **WHEN** the user runs `mzworker-go bull-run-signal` without `--symbol`
- **THEN** the command fails with a usage error and does not execute a query

### Requirement: AMQP bull-run-signal-query request type
The system SHALL accept AMQP requests of type `bull-run-signal-query` with a JSON payload `{symbol, lookbackDays}` and reply through the standard columnar reply envelope (`{requestId, hasError, value}` where `value` maps each column name to its ordered array of values).

#### Scenario: Valid request returns columnar series
- **WHEN** an AMQP request with `requestType: "bull-run-signal-query"`, a valid `symbol`, and `lookbackDays: 730` is published
- **THEN** the reply envelope has `hasError: false`
- **AND** `value` contains the weekly series columns in the query's column order

#### Scenario: Invalid request reports error
- **WHEN** an AMQP request with `requestType: "bull-run-signal-query"` and an empty `symbol` is published
- **THEN** the reply envelope has `hasError: true`

### Requirement: Lookback windowing
The system SHALL restrict the returned series to weeks within `lookbackDays` of the current date using `WHERE wk >= (current_date - lookbackDays)` in the final selection, while all window functions (rolling z-scores, moving averages, forward labels) remain computed over the full available history before trimming. The default lookback SHALL be 730 days.

#### Scenario: Default lookback spans available history
- **WHEN** the command runs with the default lookback against a dataset starting 2024-12-24
- **THEN** the earliest returned week is the first full week at or after current_date − 730 days that exists in the data

#### Scenario: Short lookback trims early weeks
- **WHEN** the command runs with `--lookback 60`
- **THEN** only weeks within the last 60 days are returned
- **AND** row values are identical to the same weeks returned by a longer lookback

### Requirement: Input validation
The system SHALL validate `BullRunSignalQuery` requests before execution: `symbol` is required and non-empty, `lookbackDays` is required and positive. Validation failures SHALL surface as errors on the CLI and as `hasError: true` on AMQP, without executing a query.

#### Scenario: Non-positive lookback rejected
- **WHEN** a request is made with `lookbackDays: 0`
- **THEN** validation fails with a descriptive error

#### Scenario: Empty symbol rejected
- **WHEN** a request is made with an empty `symbol`
- **THEN** validation fails and no query executes

### Requirement: Parity with the dynamic SQL series
The rows and columns produced by `bull-run-signal` SHALL be identical to those produced by submitting the same series SQL through `dynamic-sql-query` for the same symbol and window — same column names, same values, same week ordering.

#### Scenario: IBIT parity
- **WHEN** the command runs `-s IBIT -d 730` and the same SQL runs through `dynamic-sql-query` over the same data directory
- **THEN** both outputs contain the same number of rows with equal `wk` sequences and equal cell values
