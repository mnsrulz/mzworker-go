## ADDED Requirements

### Requirement: AMQP connection management
The system SHALL establish and maintain an AMQP 0-9-1 connection to the broker using the URL provided in the `AMQP_URL` environment variable.

#### Scenario: Successful connection
- **WHEN** the consumer starts with a valid `AMQP_URL`
- **THEN** the system opens an AMQP connection and channel successfully

#### Scenario: Invalid connection URL
- **WHEN** the consumer starts with an unreachable `AMQP_URL`
- **THEN** the system logs the error and exits with a non-zero status

### Requirement: Queue declaration
The system SHALL declare a durable request queue whose name is provided in the `AMQP_REQUEST_QUEUE` environment variable.

#### Scenario: Queue does not exist
- **WHEN** the consumer starts and the request queue does not exist
- **THEN** the system declares the queue as durable

#### Scenario: Queue already exists
- **WHEN** the consumer starts and the request queue already exists with compatible settings
- **THEN** the system uses the existing queue without error

### Requirement: Request message consumption
The system SHALL consume messages from the request queue and process each message as a JSON-encoded `ExecuteQueryRequest`.

#### Scenario: Valid request message
- **WHEN** a message arrives with valid JSON containing `query` and `limit` fields
- **THEN** the system deserializes the message and executes the query

#### Scenario: Invalid JSON payload
- **WHEN** a message arrives with malformed JSON
- **THEN** the system publishes an error response on the reply queue (if `reply_to` is set) and acknowledges the message

### Requirement: Query execution via existing logic
The system SHALL delegate query execution to the existing `executeQuery` and `enforceLimit` functions in `query.go`.

#### Scenario: Successful query
- **WHEN** a valid query is received
- **THEN** the system executes it against DuckDB and returns the result set

#### Scenario: Query validation failure
- **WHEN** the query is not a SELECT statement or contains forbidden paths
- **THEN** the system returns an error response with the appropriate error code and message

### Requirement: Response publishing via request-reply pattern
The system SHALL publish the query response as JSON-encoded `ExecuteQueryResponse` to the queue specified in the message's `reply_to` property, using the message's `correlation_id`.

#### Scenario: Response with reply_to and correlation_id
- **WHEN** a request message has `reply_to` and `correlation_id` set
- **THEN** the system publishes the response JSON to the reply queue with the matching `correlation_id`

#### Scenario: Request without reply_to
- **WHEN** a request message has no `reply_to` property
- **THEN** the system logs a warning and acknowledges the message without publishing a response

### Requirement: Message acknowledgment
The system SHALL acknowledge each message after processing (whether success or failure) to prevent redelivery.

#### Scenario: Successful processing
- **WHEN** a message is processed successfully and response is published
- **THEN** the system acknowledges the message

#### Scenario: Processing failure
- **WHEN** a message fails to process (bad JSON, query error, publish error)
- **THEN** the system acknowledges the message (after publishing error response if possible) to avoid infinite redelivery

### Requirement: Graceful shutdown
The system SHALL close the AMQP connection and channel when receiving a shutdown signal.

#### Scenario: SIGINT or SIGTERM received
- **WHEN** the process receives a termination signal
- **THEN** the consumer stops consuming, closes the AMQP channel and connection, and the process exits cleanly
