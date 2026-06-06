## Purpose
Provide deterministic Google Places Text Search geocoding behavior for Gemini-direct spot mentions, with strict Amsterdam location restriction and JSON/file debugging support.

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

### Requirement: Geocode stage supports Gemini SQLite geocoding
The `geocode-spots` stage SHALL geocode Gemini-direct spot mentions from SQLite and SHALL write geocode and article-link rows to the Gemini table set.

#### Scenario: Geocode invoked in SQLite mode
- **WHEN** a user runs `geocode-spots` in SQLite mode
- **THEN** the command SHALL read ungeocoded rows from `gemini_direct_spot_mentions`
- **AND** it SHALL write successful geocodes to `gemini_direct_spot_google_geocodes`
- **AND** it SHALL write article links to `gemini_direct_article_spots`
- **AND** it SHALL NOT require transcript-derived spot tables

### Requirement: Geocode inline export uses Gemini spot data
When `geocode-spots` performs an inline JSON export, the export step SHALL read the Gemini-direct geocode, article-link, and correction tables populated by the geocode step.

#### Scenario: Gemini geocode with inline export
- **WHEN** `geocode-spots` runs with inline export enabled
- **THEN** the geocode step SHALL write Gemini-specific geocode/link rows
- **AND** the inline export SHALL read Gemini-specific geocode/link/correction rows
- **AND** the inline export SHALL NOT read transcript-derived spot rows for that run

### Requirement: Geocode stage preserves file-mode artifact processing
The `geocode-spots` stage SHALL continue to support explicit file-mode geocoding for artifact-based processing where that mode is selected.

#### Scenario: Geocode invoked in file mode
- **WHEN** a user runs `geocode-spots --io file --in <path>`
- **THEN** the stage SHALL process the explicit input artifact
- **AND** it SHALL emit deterministic geocode output artifact(s)
- **AND** source-selected SQLite table writes SHALL NOT be required for the file-mode run

### Requirement: Gemini address hints are included in Text Search queries
The Google Places Text Search geocoding flow SHALL use a stored Gemini address as a disambiguating hint when it is available, while keeping the spot name as part of the query.

#### Scenario: Gemini mention has address
- **WHEN** the geocode stage builds a Text Search request for a Gemini spot mention with a non-empty stored address
- **THEN** the request `textQuery` SHALL be `<place>, <address>`
- **AND** the request SHALL retain the existing Amsterdam `locationRestriction`, `languageCode`, API key header, and field mask behavior
- **AND** the address SHALL NOT replace the persisted or exported spot name

#### Scenario: Gemini mention address needs city context
- **WHEN** the geocode stage builds a Text Search request for a Gemini spot mention whose stored address does not already include Amsterdam
- **THEN** the request SHALL append Amsterdam to the address portion of the query
- **AND** the appended city SHALL be used only as a Google Places lookup hint

#### Scenario: Gemini mention place plus address has no result
- **WHEN** a Gemini spot mention has an address and `<place>, <address>` returns no result
- **THEN** the geocode stage MAY retry with the address-only query
- **AND** the address-only fallback SHALL NOT replace the persisted or exported spot name

#### Scenario: Gemini mention has no address
- **WHEN** the geocode stage builds a Text Search request for a Gemini spot mention without a stored address
- **THEN** the request `textQuery` SHALL be `<place>, Amsterdam`
- **AND** the request SHALL retain the existing Amsterdam `locationRestriction`, `languageCode`, API key header, and field mask behavior

### Requirement: Geocoding persists primary type display name
The Gemini geocode stage SHALL persist the Google Places primary type display name when Google returns it.

#### Scenario: Text Search returns primary type display name
- **WHEN** Google Places Text Search returns a result with `primaryTypeDisplayName.text`
- **THEN** the stage SHALL store that text in `gemini_direct_spot_google_geocodes.primary_type_display_name`
- **AND** the stored value SHALL be available to JSON export
