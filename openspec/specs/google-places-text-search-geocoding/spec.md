## Purpose
Provide deterministic Google Places Text Search geocoding behavior with strict location restriction and a JSON-first CLI surface for scripting/debugging.

## Requirements

### Requirement: Resolve place name to coordinates via Google Places Text Search
The system SHALL expose a reusable geocoding function that accepts a place-name query string and resolves it through Google Places Text Search to a single coordinate result containing latitude and longitude.

#### Scenario: Successful coordinate resolution
- **WHEN** a non-empty place-name query is provided and the API returns at least one result within restriction
- **THEN** the function SHALL return the first valid result's latitude and longitude
- **AND** the function SHALL return no error

### Requirement: Enforce hard rectangular location restriction
Google Places Text Search requests SHALL include `locationRestriction.rectangle` and MUST NOT use location bias for this flow.

#### Scenario: Request includes required restriction bounds
- **WHEN** the geocoding function builds a Text Search request
- **THEN** it SHALL include rectangle low point latitude `52.274525` longitude `4.711585`
- **AND** it SHALL include rectangle high point latitude `52.461764` longitude `5.073559`
- **AND** it SHALL not send `locationBias`

### Requirement: Provide CLI wrapper with JSON-first output
The system SHALL provide a lightweight CLI command that accepts a place-name input and prints JSON output by default for debugging and scripting.

#### Scenario: CLI prints resolved place identity as JSON
- **WHEN** the CLI is run with a valid place-name input and a successful geocoding response
- **THEN** it SHALL print JSON containing `query`, `name`, `placeId`, and `mapsUrl`
- **AND** `mapsUrl` SHALL be derived from query + place ID using URL encoding (query encoded via standard query escaping and `placeId` percent-encoded)
- **AND** CLI success JSON SHALL NOT include latitude/longitude fields
- **AND** it SHALL exit successfully

### Requirement: Fail deterministically for invalid input and unresolved lookups
The geocoding function and CLI SHALL return explicit errors for invalid configuration, invalid input, and unresolved searches.

#### Scenario: Missing API key
- **WHEN** geocoding is invoked without required Google API credentials
- **THEN** the system SHALL fail before issuing the HTTP request
- **AND** it SHALL return an actionable configuration error

#### Scenario: Empty query
- **WHEN** geocoding is invoked with an empty or whitespace-only place-name query
- **THEN** the system SHALL reject the request
- **AND** it SHALL return a validation error indicating query input is required

#### Scenario: No results within restriction
- **WHEN** Google Places Text Search returns no candidates in the configured rectangle
- **THEN** the system SHALL return a no-result error
- **AND** the error SHALL indicate that no match was found within the enforced location restriction

#### Scenario: Upstream API failure
- **WHEN** Google Places Text Search returns a non-success HTTP/API status or malformed payload
- **THEN** the system SHALL return an error that preserves upstream failure context
- **AND** it SHALL not return fabricated coordinates

### Requirement: Geocode stage is file-mode only in this change
In this change scope, geocode stage execution SHALL support file-mode input/output handoff and SHALL reject sqlite mode.

#### Scenario: Geocode invoked in sqlite mode
- **WHEN** a user runs `geocode-spots` with `--io sqlite` (explicit or default)
- **THEN** the command SHALL fail with non-zero status
- **AND** it SHALL explain that sqlite persistence integration is deferred and file mode is required

#### Scenario: Geocode invoked in file mode
- **WHEN** a user runs `geocode-spots --io file --in <path>`
- **THEN** the stage SHALL process the explicit input artifact
- **AND** it SHALL emit deterministic geocode output artifact(s)
