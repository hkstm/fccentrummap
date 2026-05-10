# pipeline-layered-architecture Specification

## Purpose
Define and enforce layered boundaries for scraper pipeline execution (CLI → stage service → adapters) to improve maintainability, testability, and mode parity.
## Requirements
### Requirement: Pipeline architecture SHALL separate CLI, service, and adapter responsibilities
The system SHALL implement pipeline execution with explicit boundaries between CLI command handling, stage business services, and backend adapters.

#### Scenario: Stage command execution flow
- **WHEN** a user invokes any unified scrape stage command
- **THEN** the CLI layer SHALL only parse/validate invocation inputs and select runtime mode
- **AND** it SHALL delegate stage execution to a business service
- **AND** persistence and artifact operations SHALL be performed only through adapters

### Requirement: Stage services SHALL depend on stage ports, not concrete storage implementations
Each stage service SHALL define and consume narrow interfaces for the read/write operations required by that stage and SHALL NOT depend directly on concrete SQLite repository types or raw file I/O.

#### Scenario: Service dependency boundary
- **WHEN** a stage service is constructed
- **THEN** it SHALL accept interface-based dependencies for stage inputs and outputs
- **AND** replacing SQLite adapter with file adapter SHALL not require service logic changes

### Requirement: Parity-critical stages SHALL provide both SQLite and file adapters
The system SHALL provide adapter implementations for both SQLite mode and file mode for `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data`.

#### Scenario: Running parity-critical stages in either mode
- **WHEN** a parity-critical stage is executed with `--io sqlite` or `--io file`
- **THEN** the corresponding stage service SHALL execute through the selected adapter
- **AND** the stage SHALL produce contract-valid outputs for that mode

### Requirement: Stage-first package organization SHALL own contracts and orchestration while capability packages own reusable domain logic
The codebase SHALL organize pipeline internals by stage-first packages where each stage package owns stage DTOs, service orchestration, and ports. Reusable domain logic shared across stages SHALL live in explicitly named capability packages, while cross-stage generic utilities SHALL remain in dedicated shared packages.

#### Scenario: Locating stage implementation components
- **WHEN** maintainers inspect a stage package
- **THEN** they SHALL find stage-owned contracts and orchestration code in that package
- **AND** reusable non-stage-specific domain logic SHALL be located in explicit capability packages
- **AND** cross-stage shared primitives SHALL be located in `pipeline/common`

### Requirement: Stage services SHALL consume capability services through explicit boundaries
Stage orchestration in `pipeline/<stage>` SHALL depend on explicit interfaces or package-level APIs from capability packages rather than importing deprecated catch-all service packages.

#### Scenario: Wiring stage orchestration to reusable logic
- **WHEN** a stage requires reusable functionality such as content fetching, geocoding, transcription, or audio handling
- **THEN** the stage service SHALL use explicit capability package boundaries for that functionality
- **AND** it SHALL NOT introduce new dependencies on deprecated catch-all service package locations

