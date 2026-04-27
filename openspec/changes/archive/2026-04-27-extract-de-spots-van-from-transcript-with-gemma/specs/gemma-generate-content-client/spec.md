## ADDED Requirements

### Requirement: Client calls Google Generative Language generateContent endpoint
The system SHALL call Google Generative Language `generateContent` using a configured Gemma model path and JSON payload compatible with the API contract.

#### Scenario: Successful model call
- **WHEN** API credentials are configured and the endpoint responds with HTTP 2xx
- **THEN** the client SHALL return the response payload to the extraction command
- **AND** the payload SHALL be available for artifact writing and parsing

#### Scenario: Missing API key
- **WHEN** required API credentials are missing
- **THEN** the command SHALL fail before making a network request
- **AND** it SHALL emit actionable configuration guidance

### Requirement: Request payload carries Dutch prompt text as model input
The client SHALL send extraction prompt content in the request body using the API's content/parts text structure.

#### Scenario: Prompt included in request
- **WHEN** the extraction command invokes the model client
- **THEN** the request body SHALL include the composed Dutch prompt as text content
- **AND** the request SHALL target the configured Gemma model identifier

### Requirement: Non-success responses are surfaced with diagnostics
The client SHALL surface non-2xx responses as explicit command errors with status code and response body context.

#### Scenario: Endpoint returns non-2xx
- **WHEN** the model endpoint responds with HTTP 4xx or 5xx
- **THEN** the command SHALL fail with the HTTP status included
- **AND** it SHALL include enough response detail for troubleshooting
