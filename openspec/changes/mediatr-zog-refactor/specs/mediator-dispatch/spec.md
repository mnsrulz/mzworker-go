## ADDED Requirements

### Requirement: Handler registration
The system SHALL register query handlers with go-mediatr at startup, mapping request types to their corresponding handlers.

#### Scenario: Handler registered successfully
- **WHEN** the application starts and calls `mediatr.RegisterHandler` for a request type
- **THEN** the handler is available for dispatch via `mediatr.Send`

#### Scenario: Duplicate handler registration
- **WHEN** the application attempts to register two handlers for the same request type
- **THEN** the system returns an error during registration

### Requirement: Request dispatch via mediator
The system SHALL dispatch incoming requests through go-mediatr using typed `Send` calls, routing each request to its registered handler.

#### Scenario: Valid request dispatched
- **WHEN** a request is sent via `mediatr.Send` with a registered request type
- **THEN** the corresponding handler processes the request and returns a response

#### Scenario: Unknown request type
- **WHEN** a request is sent with a request type that has no registered handler
- **THEN** the system returns an error indicating the handler was not found

### Requirement: Factory registry for AMQP requests
The system SHALL maintain a factory registry mapping AMQP `requestType` strings to factory functions that create typed request structs from JSON payloads.

#### Scenario: Known request type with valid JSON
- **WHEN** an AMQP message arrives with a known `requestType` and valid JSON body
- **THEN** the factory registry creates the corresponding typed request struct

#### Scenario: Unknown request type
- **WHEN** an AMQP message arrives with an unrecognized `requestType`
- **THEN** the system logs an error and acknowledges the message without dispatching

#### Scenario: Invalid JSON payload
- **WHEN** an AMQP message arrives with valid `requestType` but malformed JSON
- **THEN** the system returns an error and publishes an error response to the reply queue

### Requirement: AMQP consumer dispatches through mediator
The system SHALL route AMQP messages through the factory registry and mediatr dispatch, replacing direct `executeQuery` calls.

#### Scenario: AMQP message processed via mediator
- **WHEN** a valid AMQP request message arrives
- **THEN** the consumer creates a typed request via the factory registry, dispatches via mediatr, and publishes the response to the reply queue

### Requirement: gRPC server dispatches through mediator
The system SHALL route gRPC `ExecuteQuery` requests through mediatr dispatch, replacing direct `executeQuery` calls.

#### Scenario: gRPC request processed via mediator
- **WHEN** a gRPC `ExecuteQuery` request arrives
- **THEN** the server creates a typed request, dispatches via mediatr, and returns the protobuf response

### Requirement: Pipeline behavior support
The system SHALL support registering pipeline behaviors that wrap all request handlers, enabling cross-cutting concerns.

#### Scenario: Pipeline behavior executes around handler
- **WHEN** a request is dispatched and a pipeline behavior is registered
- **THEN** the behavior's `Handle` method executes before and/or after the actual handler
