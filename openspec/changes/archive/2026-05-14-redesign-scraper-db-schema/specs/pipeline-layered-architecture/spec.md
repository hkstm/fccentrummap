## MODIFIED Requirements

### Requirement: Parity-critical stages SHALL provide both SQLite and file adapters
The system SHALL provide adapter implementations for both SQLite mode and file mode for `collect-article-urls`, `fetch-articles`, `extract-article-text`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data`.

#### Scenario: Running parity-critical stages in either mode
- **WHEN** a parity-critical stage is executed with `--io sqlite` or `--io file`
- **THEN** the corresponding stage service SHALL execute through the selected adapter
- **AND** the stage SHALL produce contract-valid outputs for that mode

## ADDED Requirements

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
