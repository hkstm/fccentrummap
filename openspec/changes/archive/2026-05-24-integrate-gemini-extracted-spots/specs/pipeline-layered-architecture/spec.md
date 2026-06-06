## ADDED Requirements

### Requirement: Source-consuming stage services SHALL accept a selected spot source
Stage services that consume extracted spot mentions, geocoded spot links, or spot corrections SHALL accept an explicit source selection in their request DTOs and SHALL normalize an omitted source to Gemini-direct.

#### Scenario: Geocode service receives no spot source
- **WHEN** the geocode stage service request omits the selected spot source
- **THEN** the service SHALL normalize the request to `gemini-direct`
- **AND** it SHALL pass that normalized source to its input and output ports

#### Scenario: Export service receives transcript source
- **WHEN** the export stage service request specifies `transcript`
- **THEN** the service SHALL pass `transcript` to its export input port
- **AND** it SHALL preserve the existing export assembly behavior for transcript-derived rows

#### Scenario: Correction service receives unsupported source
- **WHEN** the correction service request specifies a spot source outside `gemini-direct` or `transcript`
- **THEN** service-level validation SHALL fail before adapter writes occur
- **AND** the error SHALL name the supported source values

### Requirement: Source selection SHALL remain isolated in stage ports and adapters
The service layer SHALL orchestrate source-selected stages through source-neutral DTOs while adapters route persistence to source-specific tables.

#### Scenario: Geocode service runs for Gemini source
- **WHEN** `geocode-spots` runs with the selected source `gemini-direct`
- **THEN** the service SHALL call a stage port to list source-neutral spot mentions for geocoding
- **AND** the SQLite adapter SHALL satisfy that port by reading `gemini_direct_spot_mentions`
- **AND** the service SHALL NOT contain raw table-name branching for Gemini-specific SQL

#### Scenario: Export service runs for either source
- **WHEN** `export-data` runs with any supported selected source
- **THEN** the service SHALL request source-neutral export rows through an adapter port
- **AND** the SQLite adapter SHALL choose transcript or Gemini SQL based on the selected source
- **AND** the JSON assembly logic SHALL be shared across sources

### Requirement: Stage adapters SHALL preserve source-specific table write ownership
Source-selected adapters SHALL write only the tables owned by the invoked stage and selected spot source.

#### Scenario: Gemini-direct extraction writes accepted candidates
- **WHEN** the Gemini-direct extraction stage persists accepted candidates
- **THEN** its SQLite adapter SHALL write `gemini_direct_spot_mentions` and any presenter attribution rows required for the same article context
- **AND** it SHALL NOT write source-specific geocode, article-spot link, or correction rows

#### Scenario: Geocode stage writes Gemini geocode output
- **WHEN** the geocode stage runs with selected source `gemini-direct`
- **THEN** its SQLite adapter SHALL write `gemini_direct_spot_google_geocodes` and `gemini_direct_article_spots`
- **AND** it SHALL NOT write `gemini_direct_spot_mentions` or correction rows

#### Scenario: Correction stage writes source-selected correction output
- **WHEN** a spot correction is saved with selected source `gemini-direct`
- **THEN** its adapter SHALL write `gemini_direct_spot_corrections`
- **AND** it SHALL NOT write transcript correction storage
