## MODIFIED Requirements

### Requirement: Deterministic ordering in output
The system SHALL produce deterministic ordering of exported arrays to make output stable across runs with unchanged source data. Exported spots SHALL retain stable ordering by their existing deterministic spot fields.

#### Scenario: Repeated export with unchanged data
- **WHEN** the export is run multiple times against unchanged source data
- **THEN** the resulting JSON content MUST preserve consistent array ordering

### Requirement: Valid JSON output for empty or partial datasets
The system SHALL always write valid JSON output, including cases where source data is empty or partially available.

#### Scenario: No matching rows in source data
- **WHEN** the source dataset has no exportable rows
- **THEN** the system MUST write a valid JSON document with an empty `spots` array
- **AND** the document SHALL NOT include a top-level `presenters` array

### Requirement: JSON export supports Gemini spot source
The JSON export process SHALL read exportable geocoded spots, article links, presenter attribution, and correction data from the Gemini-direct table set.

#### Scenario: Export invoked with default source
- **WHEN** a user runs `export-data`
- **THEN** the command SHALL read Gemini-derived geocode/link rows and Gemini-specific correction rows
- **AND** the generated JSON SHALL include a `spots` array without a top-level `presenters` array
- **AND** each exported spot SHALL include presenter attribution derived from stored article presenter links
- **AND** each exported spot SHALL include the primary type display name when available

### Requirement: JSON export applies source-specific spot corrections
The JSON export process SHALL apply corrections from the Gemini-direct correction table.

#### Scenario: Gemini correction overrides exported spot
- **WHEN** a Gemini-derived exportable spot has a stored correction in `gemini_direct_spot_corrections`
- **THEN** the exported spot record SHALL apply the corrected name, place, coordinates, timestamp, or hidden state according to existing correction rules

### Requirement: JSON export keeps source-neutral public identifiers
The JSON export process SHALL keep public spot identifiers source-neutral.

#### Scenario: Gemini export builds spot identifier
- **WHEN** a Gemini-derived spot is exported
- **THEN** its public spot identifier SHALL use the existing `<article_source_id>:<mention_id>` format
- **AND** the identifier SHALL NOT include a `gemini-direct` prefix or other source marker

## ADDED Requirements

### Requirement: JSON export omits top-level presenters
The JSON export process SHALL omit the top-level `presenters` array because presenter filter values are derived from spot records by the frontend.

#### Scenario: Export contains presenters on spots only
- **WHEN** export-data writes the frontend JSON document
- **THEN** the JSON SHALL contain a top-level `spots` array
- **AND** the JSON SHALL NOT contain a top-level `presenters` array
- **AND** presenter names SHALL remain available on exported spot records for frontend derivation

### Requirement: Exported spots include primary type display name
Each exported spot record SHALL include the Google Places primary type display name when it was captured during geocoding.

#### Scenario: Primary type display name is stored
- **WHEN** an exportable Gemini geocode row has `primary_type_display_name`
- **THEN** the exported spot SHALL include that display name in the spot record

#### Scenario: Primary type display name is absent
- **WHEN** an exportable Gemini geocode row has no `primary_type_display_name`
- **THEN** the exported spot SHALL remain exportable with a null or omitted primary type display name field
