## MODIFIED Requirements

### Requirement: Unified scrape CLI provides stage subcommands
The system SHALL provide a single urfave/cli v3 entrypoint for pipeline execution with subcommands: `init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data`.

#### Scenario: List and run unified stages
- **WHEN** a user invokes the unified scrape CLI
- **THEN** the listed stage subcommands SHALL be available under that single entrypoint
- **AND** each stage subcommand SHALL delegate execution to a stage business service rather than embedding persistence-specific orchestration in the command handler

### Requirement: Unsupported stage/mode combinations fail explicitly
The unified CLI SHALL return a non-zero error with actionable guidance when a stage is invoked in an unsupported I/O mode.

#### Scenario: Unsupported stage/mode requested
- **WHEN** a user runs any stage with an unsupported `--io` mode
- **THEN** the CLI SHALL fail explicitly before stage mutations begin
- **AND** it SHALL instruct the user on the supported mode and required inputs for that stage
