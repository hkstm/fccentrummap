## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Gemini address hints are included in Text Search queries
The Google Places Text Search geocoding flow SHALL use a stored Gemini address as a disambiguating hint when it is available, while keeping the spot name as part of the query.

#### Scenario: Gemini mention has address
- **WHEN** the geocode stage builds a Text Search request for a Gemini spot mention with a non-empty stored address
- **THEN** the request `textQuery` SHALL be `<place>, <address>`
- **AND** the request SHALL retain the existing Amsterdam `locationRestriction`, `languageCode`, API key header, and field mask behavior
- **AND** the address SHALL NOT replace the persisted or exported spot name

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
