## ADDED Requirements

### Requirement: Zog schema definition for request types
The system SHALL define Zog schemas for each request type, with schema keys matching PascalCase struct field names.

#### Scenario: Schema defined with correct field mappings
- **WHEN** a Zog schema is defined for a request type
- **THEN** schema keys match the Go struct field names (PascalCase) and validation rules are applied to each field

### Requirement: Request validation at handler entry
The system SHALL validate incoming requests using Zog at the top of each handler's `Handle` method before executing business logic.

#### Scenario: Valid request passes validation
- **WHEN** a handler receives a request that satisfies all schema rules
- **THEN** the handler proceeds to business logic without error

#### Scenario: Invalid request fails validation
- **WHEN** a handler receives a request that violates a schema rule (e.g., missing required field, invalid format)
- **THEN** the handler returns a validation error with detailed issue information and does NOT execute business logic

### Requirement: JSON parsing via zjson
The system SHALL use `zjson.Decode` to parse JSON payloads into typed request structs, combining unmarshaling and validation in a single step.

#### Scenario: JSON parsed and validated successfully
- **WHEN** a valid JSON payload is parsed via `zjson.Decode` into a struct
- **THEN** the struct is populated with parsed values and all validation rules pass

#### Scenario: Invalid JSON input
- **WHEN** a malformed JSON payload is provided to `zjson.Decode`
- **THEN** a top-level ZogIssue with `IssueCodeInvalidJSON` is returned

### Requirement: Validation error response format
The system SHALL return validation errors as structured data containing field-level issue details.

#### Scenario: Validation error contains field details
- **WHEN** validation fails on multiple fields
- **THEN** the error contains a map of field names to lists of issue messages (via `z.Issues.SanitizeMap`)
