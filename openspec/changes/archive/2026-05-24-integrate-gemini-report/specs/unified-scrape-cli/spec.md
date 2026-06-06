## MODIFIED Requirements

### Requirement: Unified scrape CLI provides stage subcommands
The system SHALL provide a single urfave/cli v3 entrypoint for pipeline execution with subcommands: `init`, `collect-article-urls`, `fetch-articles`, `extract-article-text`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `extract-spots-gemini-direct`, `compare-spot-extractions`, `geocode-spots`, and `export-data`.

#### Scenario: List and run unified stages
- **WHEN** a user invokes the unified scrape CLI
- **THEN** the listed stage subcommands SHALL be available under that single entrypoint
- **AND** each stage subcommand SHALL delegate execution to a stage business service rather than embedding persistence-specific orchestration in the command handler

#### Scenario: Experimental Gemini-direct commands are explicit
- **WHEN** a user lists or invokes the unified scrape CLI stages
- **THEN** the Gemini-direct extraction and comparison report commands SHALL be exposed as explicit commands separate from `extract-spots`, `geocode-spots`, and `export-data`
- **AND** existing canonical stage commands SHALL retain their current default behavior
