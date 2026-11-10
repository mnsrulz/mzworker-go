## ADDED Requirements

### Requirement: Level-based logging configuration
The system SHALL use leveled logging with a default minimum level of `Info`. The CLI SHALL provide a persistent `-v` / `--verbose` flag, defaulting to `false`, that raises the minimum level to `Debug`. The system SHALL also accept a `LOG_LEVEL` environment variable with the values `debug`, `info`, `warn`, or `error` (case-insensitive).

#### Scenario: Default level is Info
- **WHEN** a command runs without `-v` / `--verbose` and `LOG_LEVEL` is unset
- **THEN** the minimum log level is `Info` and debug records are not emitted

#### Scenario: Verbose flag raises the level to Debug
- **WHEN** a user runs any command with `-v` or `--verbose`
- **THEN** the minimum log level is `Debug` and debug records are emitted

#### Scenario: Environment sets the level
- **WHEN** a command runs with `LOG_LEVEL=debug` and no verbose flag
- **THEN** the minimum log level is `Debug`

#### Scenario: Invalid environment value falls back
- **WHEN** a command runs with an unparsable `LOG_LEVEL` value
- **THEN** the minimum level remains `Info` and a warning record is logged

### Requirement: Executed SQL logged at Debug
The system SHALL log the full SQL statement sent to DuckDB — including the CTE scaffold, the command SQL, and the `LIMIT` clause — at `Debug`, to stderr, before the query executes. The SQL SHALL NOT be logged at the default `Info` level.

#### Scenario: SQL logged when verbose is enabled
- **WHEN** a query command runs with `-v` or `LOG_LEVEL=debug`
- **THEN** the full composed SQL sent to DuckDB is logged to stderr at `Debug`

#### Scenario: No SQL logged by default
- **WHEN** a query command runs at the default `Info` level
- **THEN** no executed-SQL record is logged

#### Scenario: Logged SQL matches the executed statement
- **WHEN** the SQL is logged and the query executes
- **THEN** the logged statement is the same string used for `db.QueryContext`, including the appended `LIMIT`

### Requirement: Leveled logging preserves output
All log records SHALL be written to stderr and SHALL NOT change query results, output formatting, error values, or AMQP reply payloads.

#### Scenario: Machine-readable stdout stays clean
- **WHEN** a user runs a query command with `-v` and `--format json`
- **THEN** stdout remains valid JSON with no log text mixed in

#### Scenario: Results unchanged by verbosity
- **WHEN** the same query runs with and without `-v`
- **THEN** the returned columns and rows are identical

### Requirement: All query paths are covered
The executed-SQL record SHALL be emitted for every path that executes SQL, including the dynamic `query` subcommand and the shortcut commands.

#### Scenario: Shortcut command logs its query
- **WHEN** a user runs a shortcut command such as `ohlc` or `options-stat` with `-v`
- **THEN** the full composed SQL for that command is logged at `Debug`

#### Scenario: Dynamic query logs its query
- **WHEN** a user runs the `query` subcommand with `-v`
- **THEN** the full composed SQL, including the CTE scaffold wrapping the user SQL, is logged at `Debug`

### Requirement: Existing records keep default visibility
The migration to leveled logging SHALL preserve the default visibility of existing lifecycle, timing, validation, and failure records: informational records SHALL be logged at `Info` or a higher severity, and failure records SHALL be logged at `Error` or `Warn`.

#### Scenario: Lifecycle records visible by default
- **WHEN** the daemon starts or stops at the default `Info` level
- **THEN** the existing lifecycle messages are still emitted

#### Scenario: Failure records visible by default
- **WHEN** a query or AMQP dispatch fails at the default `Info` level
- **THEN** the failure is still logged, at `Error` or `Warn`

### Requirement: Daemon supports verbosity via environment
The `serve` command SHALL enable the executed-SQL record when `LOG_LEVEL=debug` is set, since it takes no verbose CLI flag.

#### Scenario: Logging enabled by environment
- **WHEN** `serve` runs with `LOG_LEVEL=debug`
- **THEN** each query processed from the AMQP queue logs its full SQL to stderr at `Debug`

#### Scenario: Logging disabled without flag or environment
- **WHEN** `serve` runs with `LOG_LEVEL` unset
- **THEN** no executed-SQL records are logged
