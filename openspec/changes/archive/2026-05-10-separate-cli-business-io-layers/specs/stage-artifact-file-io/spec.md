## MODIFIED Requirements

### Requirement: File mode requires explicit input path
File mode execution SHALL require explicit `--in` path(s) for stage inputs.

#### Scenario: Missing required file input
- **WHEN** a stage is run with `--io file` and required `--in` input is omitted
- **THEN** the stage SHALL fail with a non-zero validation error

### Requirement: File outputs use deterministic identity-based naming
File-mode outputs SHALL be derived deterministically from upstream identity plus stage/type suffixes.

#### Scenario: Derive chained artifact name
- **WHEN** a stage processes an artifact with a stable identity key
- **THEN** the stage SHALL emit output artifact names that preserve that identity key
- **AND** it SHALL append stage/type suffixes consistently

### Requirement: No implicit latest-artifact discovery
The system SHALL NOT auto-select latest artifacts in file mode.

#### Scenario: User omits explicit artifact reference
- **WHEN** a stage needs file input and no explicit input path is provided
- **THEN** the stage SHALL fail instead of selecting a latest artifact automatically

## ADDED Requirements

### Requirement: File-mode stage execution SHALL use typed stage contracts
File-mode execution for `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data` SHALL parse and produce typed stage payloads instead of passthrough copies.

#### Scenario: File-mode stage processes contract-valid input
- **WHEN** a user runs a parity-critical stage in file mode with a valid input artifact
- **THEN** the stage SHALL load and validate the stage-specific payload contract
- **AND** it SHALL execute stage business logic
- **AND** it SHALL emit a stage-specific output payload contract

### Requirement: File artifacts SHALL prioritize debugging clarity
File artifact payloads SHALL remain simple and human-inspectable for ad-hoc debugging and SHALL NOT require explicit schema version metadata.

#### Scenario: Maintainer inspects stage artifacts
- **WHEN** a maintainer opens a generated stage artifact file
- **THEN** the payload SHALL be directly readable and useful for debugging stage handoff behavior
- **AND** the payload SHALL not include schema-version metadata fields
