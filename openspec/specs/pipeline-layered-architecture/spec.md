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
The system SHALL provide adapter implementations for both SQLite mode and file mode for `collect-article-urls`, `fetch-articles`, `extract-article-text`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data`.

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

### Requirement: Stage adapters SHALL enforce single-writer table ownership
Pipeline adapters SHALL implement writes so each persistent table is mutated by only its designated writer stage.

#### Scenario: Stage adapter persistence boundaries
- **WHEN** maintainers inspect adapter write paths across stages
- **THEN** `extract-spots` SHALL be the only writer for `spot_mentions`, `presenters`, and `article_presenters`
- **AND** `geocode-spots` SHALL be the only writer for `spot_google_geocodes` and `article_spots`

### Requirement: Extract-spots stage SHALL materialize presenter linkage
The extract-spots orchestration SHALL persist extracted spot mentions and presenter linkage in one stage flow through stage-owned adapter writes.

#### Scenario: Extract-spots returns presenter_name and places
- **WHEN** extraction output contains a presenter and one or more places
- **THEN** the stage SHALL write `spot_mentions` rows for places
- **AND** it SHALL upsert `presenters` and link rows in `article_presenters` for the article context

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
