## ADDED Requirements

### Requirement: SQLite schema SHALL store article text extraction outcomes and content
The repository schema SHALL include dedicated tables for article text extraction outcomes and extracted text segments linked to raw articles.

#### Scenario: Fresh database initialization includes extraction tables
- **WHEN** the repository initializes schema on a fresh database
- **THEN** it SHALL create a table for extraction outcomes linked to `article_raw_id`
- **AND** it SHALL create a table for extracted text segments linked to both `article_raw_id` and extraction outcome

#### Scenario: Existing database initialization remains idempotent
- **WHEN** the repository initializes schema on an existing database
- **THEN** it SHALL create any missing extraction schema objects without dropping existing data
- **AND** repeated initialization SHALL remain idempotent

### Requirement: Repository writes SHALL replace prior extraction result per article atomically
The repository SHALL replace prior extraction records for an article with one authoritative latest outcome in a single transaction.

#### Scenario: Persist successful matched extraction
- **WHEN** a matched extraction result is saved for an article
- **THEN** the repository SHALL replace prior extraction records for that article in one transaction
- **AND** it SHALL persist one extraction outcome row plus associated extracted text segment rows

#### Scenario: Persist no-match extraction
- **WHEN** a no-match extraction result is saved for an article
- **THEN** the repository SHALL persist one extraction outcome row with status `no_match`
- **AND** it SHALL persist zero extracted text segment rows for that article

#### Scenario: Persist extraction error outcome
- **WHEN** an extraction error outcome is saved for an article
- **THEN** the repository SHALL persist one extraction outcome row with status `error`
- **AND** it SHALL persist error diagnostics without requiring extracted text segment rows
