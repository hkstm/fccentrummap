## Purpose
Define a single stage-based CLI entrypoint for scraper pipeline execution with explicit I/O mode behavior and preflight validation.

## Requirements

### Requirement: Unified scrape CLI provides stage subcommands
The system SHALL provide a single CLI entrypoint for pipeline execution with subcommands: `init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data`.

#### Scenario: List and run unified stages
- **WHEN** a user invokes the unified scrape CLI
- **THEN** the listed stage subcommands SHALL be available under that single entrypoint

### Requirement: Unsupported stage/mode combinations fail explicitly
The unified CLI SHALL return a non-zero error with actionable guidance when a stage is invoked in an unsupported I/O mode.

#### Scenario: Unsupported geocode sqlite mode
- **WHEN** a user runs `geocode-spots` with default `--io sqlite`
- **THEN** the CLI SHALL fail explicitly
- **AND** it SHALL instruct the user to use the currently supported mode (file mode with explicit `--in`)

### Requirement: Init performs API preflight validation
The `init` stage SHALL validate required API credentials before starting pipeline setup.

#### Scenario: Missing required API credentials
- **WHEN** one or more required API credentials are missing
- **THEN** `init` SHALL fail with non-zero status
- **AND** it SHALL report which required environment variables are missing
