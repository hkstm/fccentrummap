## MODIFIED Requirements

### Requirement: Unified scrape CLI provides Gemini-first stage subcommands
The system SHALL provide a single urfave/cli v3 entrypoint for pipeline execution with subcommands for the current Gemini-first pipeline: `init`, `collect-article-urls`, `fetch-articles`, `extract-spots-gemini-direct`, `geocode-spots`, and `export-data`.

#### Scenario: List and run unified stages
- **WHEN** a user invokes the unified scrape CLI
- **THEN** the listed current stage subcommands SHALL be available under that single entrypoint
- **AND** each stage subcommand SHALL delegate execution to a stage business service rather than embedding persistence-specific orchestration in the command handler
- **AND** legacy transcription-based stages SHALL NOT be required for normal pipeline execution

#### Scenario: Gemini-direct extraction command is available
- **WHEN** a user lists or invokes the unified scrape CLI stages
- **THEN** the Gemini-direct extraction command SHALL be exposed as the extraction command for the current data pipeline
- **AND** downstream `geocode-spots` and `export-data` commands SHALL consume Gemini-direct data by default

### Requirement: Geocode stage SHALL materialize Gemini spot links in the same command
The `geocode-spots` command SHALL persist Gemini geocode results and article-to-spot link rows in one stage execution.

#### Scenario: Geocode command completes successfully
- **WHEN** `geocode-spots` resolves coordinates for extracted Gemini mentions
- **THEN** it SHALL write `gemini_direct_spot_google_geocodes` rows for mentions
- **AND** it SHALL write corresponding `gemini_direct_article_spots` rows in the same run

## REMOVED Requirements

### Requirement: Legacy transcription stage commands remain part of the current pipeline
**Reason**: The Gemini-direct flow is now the source of truth and the old transcription-based extraction flow is being removed.
**Migration**: Use `extract-spots-gemini-direct` followed by `geocode-spots` and `export-data` for the current pipeline. Historical command behavior remains available only from version-control history, not the active CLI contract.
