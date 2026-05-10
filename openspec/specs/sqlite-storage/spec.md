## Purpose

Define the canonical SQLite schema and repository-layer behavior used by the current scraper and exporter.
## Requirements
### Requirement: Initialize schema on startup
The scraper repository layer SHALL create required tables if they do not already exist.

#### Scenario: Fresh database
- **WHEN** the program opens a new `data/spots.db`
- **THEN** it SHALL create `articles_raw`, `authors`, `spots`, `articles`, `article_spots`, `article_audio_sources`, and `article_audio_transcriptions`
- **AND** it SHALL enforce uniqueness for transcription rows on `(audio_source_id, provider, language)`

#### Scenario: Existing database
- **WHEN** the program opens an existing database with the required schema
- **THEN** existing data SHALL be preserved
- **AND** missing newly required transcription schema objects SHALL be created idempotently

### Requirement: Store raw article HTML as pending work
The repository SHALL store fetched article HTML in `articles_raw` with status `PENDING`.

#### Scenario: New raw article
- **WHEN** a new article URL and HTML are inserted
- **THEN** the repository SHALL create an `articles_raw` row with status `PENDING`

#### Scenario: Duplicate raw article URL
- **WHEN** an article URL already exists in `articles_raw`
- **THEN** the insert SHALL be skipped without error

### Requirement: Query pending articles
The repository SHALL provide access to pending raw articles for downstream processing.

#### Scenario: Pending articles exist
- **WHEN** rows with status `PENDING` exist in `articles_raw`
- **THEN** the repository SHALL return their IDs, URLs, and HTML

### Requirement: Export query supports frontend JSON generation
The repository SHALL provide a query that joins spots, articles, and authors into export-ready data.

#### Scenario: Exporting current map data
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return deduplicated spot records with their coordinates and associated author names

### Requirement: Failure status updates are logged
The repository SHALL record failure status changes and log the failure reason.

#### Scenario: Marking a raw article as failed
- **WHEN** article processing fails and the repository updates an `articles_raw` row to `FAILED`
- **THEN** it SHALL refresh `updated_at`
- **AND** it SHALL log the article ID, URL, and failure reason

### Requirement: Store per-article audio as SQLite blobs
The repository SHALL support durable storage of downloaded audio payloads for article-linked videos.

#### Scenario: Insert acquired audio
- **WHEN** audio is acquired for an article-linked video
- **THEN** the repository SHALL store an `article_audio_sources` row with `article_raw_id`, `video_id`, `youtube_url`, `audio_format`, `mime_type`, `byte_size`, and `audio_blob`

#### Scenario: Duplicate audio for same article
- **WHEN** an `article_audio_sources` row already exists for an `article_raw_id`
- **THEN** the pipeline SHALL skip duplicate insertion unless explicitly forced

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

### Requirement: Stage-mode support matrix is enforced by command-layer validation
The system SHALL enforce declared stage/mode support and reject unsupported combinations before processing.

#### Scenario: Unsupported stage/mode requested
- **WHEN** a stage is requested with an unsupported I/O mode
- **THEN** validation SHALL fail before any data mutation
- **AND** the command SHALL return a non-zero error with guidance

### Requirement: Current change defers new geocode-to-final-table SQLite writes
This change SHALL NOT require introducing new SQLite write paths from geocode stage into final export tables.

#### Scenario: Geocode stage executed in current scope
- **WHEN** geocode stage is executed in this change scope
- **THEN** SQLite final-table writes from geocode SHALL be treated as deferred
- **AND** file artifact handoff SHALL remain the temporary supported contract

### Requirement: SQLite backend SHALL act as an adapter behind stage ports
SQLite-backed persistence SHALL be accessed through stage adapter implementations that satisfy stage port contracts used by business services.

#### Scenario: Stage service executes in sqlite mode
- **WHEN** a parity-critical stage is run with `--io sqlite`
- **THEN** the stage service SHALL call SQLite adapter implementations through stage ports
- **AND** the service layer SHALL remain independent of concrete SQLite repository types

### Requirement: SQLite adapter SHALL preserve integrity constraints as canonical guarantees
SQLite adapter behavior SHALL preserve canonical integrity guarantees defined by schema constraints and transactional repository behavior.

#### Scenario: SQLite adapter persists linked stage data
- **WHEN** SQLite adapter writes related records for a stage flow
- **THEN** foreign-key and uniqueness guarantees SHALL be enforced by SQLite schema constraints
- **AND** constraint violations SHALL surface as non-zero stage failures with actionable errors

