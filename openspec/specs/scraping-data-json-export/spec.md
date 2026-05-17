# scraping-data-json-export Specification

## Purpose
TBD - created by archiving change export-scraping-data-to-json. Update Purpose after archive.
## Requirements
### Requirement: CLI-triggered JSON export
The system SHALL provide a CLI option to generate scraping data JSON export as an optional operation.

#### Scenario: User invokes export option
- **WHEN** the user executes the scraping CLI with the export option enabled
- **THEN** the system MUST run the JSON export process

### Requirement: Configurable export output path
The system SHALL allow users to specify the output path for the generated JSON export.

#### Scenario: User provides output path
- **WHEN** the user passes an explicit export output path
- **THEN** the system MUST write the JSON export to that path

### Requirement: Deterministic ordering in output
The system SHALL produce deterministic ordering of exported arrays to make output stable across runs with unchanged source data. Exported spots SHALL retain stable ordering by their existing deterministic spot fields, and exported presenters SHALL be ordered by latest associated article publication time descending with `presenterName` ascending as the tie-breaker.

#### Scenario: Repeated export with unchanged data
- **WHEN** the export is run multiple times against unchanged source data
- **THEN** the resulting JSON content MUST preserve consistent array ordering

#### Scenario: Presenter ordering uses stored article publish metadata
- **WHEN** exported presenters have associated articles with stored publication times
- **THEN** the export SHALL sort the `presenters` array by each presenter's latest associated article publication time from newest to oldest

#### Scenario: Presenter ordering uses deterministic tie-breakers
- **WHEN** multiple exported presenters have the same latest associated article publication time
- **THEN** those presenters SHALL sort by `presenterName` ascending

#### Scenario: Export fails without required publish metadata
- **WHEN** one or more exported presenters are associated with exportable articles that have no stored publication time
- **THEN** the export SHALL fail with an actionable error instead of producing JSON with fallback ordering

### Requirement: Valid JSON output for empty or partial datasets
The system SHALL always write valid JSON output, including cases where source data is empty or partially available.

#### Scenario: No matching rows in source data
- **WHEN** the source dataset has no exportable rows
- **THEN** the system MUST write a valid JSON document with empty `spots` and `presenters` arrays

### Requirement: JSON export applies spot corrections
The JSON export process SHALL apply stored spot corrections to exported spot records while preserving existing output compatibility for consumers.

#### Scenario: Name correction overrides exported spot name
- **WHEN** an exportable spot has a stored spot name correction
- **THEN** the exported spot record SHALL use the corrected `spotName`

#### Scenario: Place correction overrides exported place and coordinates
- **WHEN** an exportable spot has a stored Google place ID correction with resolved coordinates
- **THEN** the exported spot record SHALL use the corrected `placeId`, `latitude`, and `longitude`

#### Scenario: Timestamp correction overrides exported YouTube timestamp
- **WHEN** an exportable spot has a stored timestamp correction extracted from a timestamped `youtube.com` or `youtu.be` URL
- **THEN** the exported `youtubeLink` SHALL point to the same source video with the corrected timestamp

#### Scenario: Hidden correction omits spot
- **WHEN** an exportable spot has a stored hidden correction
- **THEN** the exported `spots` array SHALL omit that spot

#### Scenario: Empty correction fields preserve source values
- **WHEN** a correction exists but a supported override field is empty or null
- **THEN** the exported spot record SHALL use the original source value for that field

### Requirement: Correction export failures are actionable
The JSON export process SHALL fail instead of producing partially corrected output when required correction data is invalid.

#### Scenario: Corrected place ID lacks coordinates
- **WHEN** a spot correction contains a corrected place ID without resolved corrected coordinates
- **THEN** the export SHALL fail with an actionable error naming the affected stable spot identifier

#### Scenario: Corrected timestamp is invalid
- **WHEN** a spot correction contains an invalid timestamp value that was not extracted from a full timestamped `youtube.com` or `youtu.be` URL
- **THEN** the export SHALL fail with an actionable error naming the affected stable spot identifier

