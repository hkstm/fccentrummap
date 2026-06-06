## ADDED Requirements

### Requirement: JSON export supports selected spot source
The JSON export process SHALL read exportable geocoded spots, article links, presenter attribution, and correction data from the selected spot source while preserving the existing output schema.

#### Scenario: Export invoked with default source
- **WHEN** a user runs `export-data` without `--spot-source`
- **THEN** the command SHALL select `gemini-direct`
- **AND** the export query SHALL read Gemini-derived geocode/link rows and Gemini-specific correction rows
- **AND** the generated JSON SHALL preserve the existing frontend-compatible shape

#### Scenario: Export invoked with transcript source
- **WHEN** a user runs `export-data --spot-source transcript`
- **THEN** the export query SHALL read transcript-derived geocode/link rows and transcript correction rows
- **AND** the generated JSON SHALL match the existing transcript-derived export behavior
- **AND** it SHALL NOT read Gemini-specific correction rows

#### Scenario: Export invoked with unsupported source
- **WHEN** a user runs `export-data` with an unsupported `--spot-source` value
- **THEN** validation SHALL fail before writing the output file
- **AND** the error SHALL name `gemini-direct` and `transcript` as supported values

### Requirement: JSON export applies source-specific spot corrections
The JSON export process SHALL apply corrections from the correction table associated with the selected spot source and SHALL ignore correction rows from unselected sources.

#### Scenario: Gemini correction overrides exported spot
- **WHEN** a Gemini-derived exportable spot has a stored correction in `gemini_direct_spot_corrections`
- **THEN** the exported spot record SHALL apply the corrected name, place, coordinates, timestamp, or hidden state according to existing correction rules
- **AND** transcript-derived `spot_corrections` rows with the same public spot identifier SHALL NOT affect the Gemini export

#### Scenario: Transcript correction overrides exported spot
- **WHEN** a transcript-derived exportable spot has a stored correction in `spot_corrections`
- **THEN** the exported spot record SHALL apply the corrected name, place, coordinates, timestamp, or hidden state according to existing correction rules
- **AND** Gemini-specific correction rows with the same public spot identifier SHALL NOT affect the transcript export

### Requirement: JSON export keeps source-neutral public identifiers
The JSON export process SHALL keep public spot identifiers source-neutral and SHALL scope identifier lookup by the selected spot source.

#### Scenario: Gemini export builds spot identifier
- **WHEN** a Gemini-derived spot is exported
- **THEN** its public spot identifier SHALL use the existing `<article_source_id>:<mention_id>` format
- **AND** the identifier SHALL NOT include a `gemini-direct` prefix or other source marker

#### Scenario: Correction lookup uses selected source
- **WHEN** the export process resolves corrections for a public spot identifier
- **THEN** it SHALL look up corrections only in the selected source's correction storage
- **AND** same-string identifiers in the unselected source SHALL NOT affect the export
