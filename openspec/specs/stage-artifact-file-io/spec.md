## Purpose
Define deterministic file-mode handoff rules for stage input/output artifacts across the unified scrape pipeline.

## Requirements

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
