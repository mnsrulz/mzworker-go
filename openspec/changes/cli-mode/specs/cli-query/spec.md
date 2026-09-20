## ADDED Requirements

### Requirement: CLI accepts SQL from command line argument
The system SHALL accept raw SQL as a command line argument via `mzworker-go <SQL>` or `mzworker-go query <SQL>`.

#### Scenario: SQL as first argument
- **WHEN** user runs `mzworker-go "SELECT 42 AS answer"`
- **THEN** system executes the SQL query against DuckDB and outputs results

#### Scenario: SQL with query subcommand
- **WHEN** user runs `mzworker-go query "SELECT 42 AS answer"`
- **THEN** system executes the SQL query against DuckDB and outputs results

### Requirement: CLI accepts SQL from stdin
The system SHALL accept raw SQL from stdin when no SQL argument is provided.

#### Scenario: Piped SQL
- **WHEN** user runs `echo "SELECT 42 AS answer" | mzworker-go query`
- **THEN** system reads SQL from stdin and executes it against DuckDB

### Requirement: JSON output format
The system SHALL output results as a flat JSON array of objects by default.

#### Scenario: Default JSON output
- **WHEN** user runs `mzworker-go "SELECT 42 AS answer"`
- **THEN** system outputs `[{"answer": 42}]`

#### Scenario: Explicit JSON format
- **WHEN** user runs `mzworker-go query "SELECT 42 AS answer" --format json`
- **THEN** system outputs `[{"answer": 42}]`

### Requirement: Table output format
The system SHALL output results as an ASCII table when `--format table` is specified.

#### Scenario: Table format
- **WHEN** user runs `mzworker-go query "SELECT 42 AS answer" --format table`
- **THEN** system outputs an ASCII table with column headers and aligned values

### Requirement: Raw output format
The system SHALL output results in nested `{columns, rows}` format when `--format raw` is specified.

#### Scenario: Raw format
- **WHEN** user runs `mzworker-go query "SELECT 42 AS answer" --format raw`
- **THEN** system outputs `{"columns": ["answer"], "rows": [[42]]}`

### Requirement: Limit flag
The system SHALL accept a `--limit` flag to override the default row limit.

#### Scenario: Custom limit
- **WHEN** user runs `mzworker-go query "SELECT * FROM trades" --limit 500`
- **THEN** system executes the query with LIMIT 500 applied (if not already present in SQL)

### Requirement: Serve subcommand starts AMQP consumer
The system SHALL start the AMQP consumer when the `serve` subcommand is used.

#### Scenario: AMQP daemon mode
- **WHEN** user runs `mzworker-go serve`
- **AND** environment variables `AMQP_URL` and `AMQP_REQUEST_QUEUE` are set
- **THEN** system starts consuming from the AMQP queue

### Requirement: Unknown subcommand treated as SQL
The system SHALL treat an unrecognized first argument as SQL rather than erroring.

#### Scenario: Bare SQL without subcommand
- **WHEN** user runs `mzworker-go "SELECT 1"`
- **THEN** system executes the SQL query (same as `mzworker-go query "SELECT 1"`)

### Requirement: No arguments shows usage
The system SHALL display usage information when no arguments and no stdin are provided.

#### Scenario: No arguments
- **WHEN** user runs `mzworker-go` with no arguments and stdin is a terminal
- **THEN** system displays usage information and exits
