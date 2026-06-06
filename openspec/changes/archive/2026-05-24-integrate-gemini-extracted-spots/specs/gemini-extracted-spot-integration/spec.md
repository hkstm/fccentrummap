## ADDED Requirements

### Requirement: Gemini-derived spots persist as selectable database source
The system SHALL treat successful Gemini-direct extraction output as a first-class spot source by persisting valid candidates to SQLite and making them selectable by downstream spot-consuming stages without requiring transcript rows.

#### Scenario: Valid Gemini response is accepted
- **WHEN** `extract-spots-gemini-direct` parses a valid response for an article source and YouTube URL in SQLite mode
- **THEN** the system SHALL persist each valid spot candidate to `gemini_direct_spot_mentions`
- **AND** each persisted mention SHALL link directly to the `article_sources` row and store the source YouTube URL
- **AND** persistence SHALL NOT require `audio_sources` or `audio_transcriptions` rows to exist

#### Scenario: Gemini source is consumed downstream
- **WHEN** a downstream spot-consuming stage runs with `--spot-source gemini-direct` or with no explicit spot source
- **THEN** the stage SHALL read and write the Gemini-specific table set for spot mentions, geocodes, article links, and corrections
- **AND** it SHALL NOT read transcript-derived `spot_mentions` as the selected spot source

### Requirement: Transcript and Gemini spot sources remain independently selectable
The system SHALL keep transcript-derived and Gemini-derived spot data separate and SHALL require source-consuming downstream stages to operate on exactly one selected source per invocation.

#### Scenario: Default source is omitted
- **WHEN** a source-consuming downstream command is invoked without `--spot-source`
- **THEN** the command SHALL normalize the selected source to `gemini-direct`
- **AND** command output or diagnostics SHALL identify Gemini-direct as the selected source

#### Scenario: Transcript fallback is requested
- **WHEN** a source-consuming downstream command is invoked with `--spot-source transcript`
- **THEN** the command SHALL read and write the existing transcript-derived table set
- **AND** it SHALL NOT mutate Gemini-specific spot, geocode, article-link, or correction rows

#### Scenario: Unsupported spot source is requested
- **WHEN** a source-consuming downstream command receives a spot source other than `gemini-direct` or `transcript`
- **THEN** validation SHALL fail before any source-specific data mutation
- **AND** the error SHALL name the supported spot source values

### Requirement: Export shape remains stable across spot sources
The system SHALL emit the existing frontend JSON schema regardless of whether the selected spot source is Gemini-direct or transcript-derived.

#### Scenario: Gemini-derived data is exported
- **WHEN** `export-data` exports with the selected source `gemini-direct`
- **THEN** the resulting JSON SHALL use the same top-level and spot fields as transcript-derived export output
- **AND** public spot identifiers SHALL use the existing `<article_source_id>:<mention_id>` shape without a source prefix

#### Scenario: Transcript-derived data is exported
- **WHEN** `export-data` exports with the selected source `transcript`
- **THEN** the resulting JSON SHALL preserve the existing transcript-derived output shape and identifier behavior
- **AND** it SHALL NOT include Gemini-specific provenance fields in the exported JSON
