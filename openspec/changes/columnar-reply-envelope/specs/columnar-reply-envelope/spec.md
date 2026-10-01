## ADDED Requirements

### Requirement: Reply envelope
Every AMQP reply SHALL be a JSON object with exactly the fields `requestId`, `hasError`, and `value`. `requestId` SHALL be echoed from the incoming request's `requestId` field.

#### Scenario: Successful query reply
- **WHEN** a query request with `"requestId": "d14eb4a0-be85-4034-a37f-969c90a9f024"` succeeds
- **THEN** the reply is `{"requestId": "d14eb4a0-be85-4034-a37f-969c90a9f024", "hasError": false, "value": {...}}`

#### Scenario: Request without requestId
- **WHEN** the incoming request JSON contains no `requestId` field
- **THEN** the reply carries `"requestId": ""` and remains a well-formed envelope

### Requirement: Columnar query result
For query responses, `value` SHALL be an object whose keys are the result's column names and whose values are arrays holding that column's values for each row, in row order. The key order SHALL match the column order returned by the query.

#### Scenario: Multi-column result
- **WHEN** a query returns columns `dt, last_close, straddle_price, expiry` with N rows
- **THEN** `value` is `{"dt": [N values], "last_close": [N values], "straddle_price": [N values], "expiry": [N values]}` in that key order, and each array has length N

#### Scenario: Value types are preserved
- **WHEN** a column holds SQL numbers, strings, or booleans
- **THEN** each array element is serialized as the corresponding JSON number, string, or boolean — not a wrapped object

### Requirement: Null handling
SQL `NULL` values SHALL be serialized as JSON `null` in the column array; array length SHALL equal the row count regardless of nulls.

#### Scenario: Null cell
- **WHEN** a cell in row 2 of column `x` is NULL
- **THEN** `value.x[1]` is `null` (the element is present, not skipped)

### Requirement: Error reply
When request parsing, dispatch, or response conversion fails, the reply SHALL set `hasError` to `true` and `value` to an empty object — the same shape the reference consumer publishes for failures (`publish(requestId, true, {})`). The failure message SHALL be written to the worker log, not to the reply.

#### Scenario: Dispatch failure
- **WHEN** handler dispatch returns an error for a request with a reply-to
- **THEN** the reply is `{"requestId": "...", "hasError": true, "value": {}}` and the error message appears in the log

### Requirement: Non-query responses
For non-query response types, `value` SHALL carry the handler's response payload as-is.

#### Scenario: Ping response
- **WHEN** a ping request succeeds
- **THEN** `value` is the ping payload object (e.g. `{"message": ..., "server_time": ...}`) inside the same envelope
