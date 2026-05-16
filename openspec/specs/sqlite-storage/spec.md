## Purpose

Define the canonical SQLite schema and repository-layer behavior used by the current scraper and exporter.
## Requirements
### Requirement: Initialize schema on startup
The scraper repository layer SHALL create required tables and compatible schema extensions if they do not already exist.

#### Scenario: Fresh database
- **WHEN** the program opens a new `data/spots.db`
- **THEN** it SHALL create `article_sources`, `article_fetches`, `article_texts`, `audio_sources`, `audio_transcriptions`, `spot_mentions`, `spot_google_geocodes`, `presenters`, `article_presenters`, and `article_spots`
- **AND** `article_sources` SHALL include nullable publication timestamp storage for article publish time
- **AND** it SHALL enforce uniqueness for transcription rows on `(audio_source_id, provider, language)`
- **AND** it SHALL enforce one-row-per-source latest fetch uniqueness on `article_fetches(article_source_id)`

#### Scenario: Existing database
- **WHEN** the program opens an existing database with the required schema
- **THEN** existing data SHALL be preserved
- **AND** missing newly required schema objects or compatible columns, including article publication timestamp storage, SHALL be created idempotently for compatible v2-era databases
- **AND** legacy pre-v2 database files are unsupported and SHALL be reinitialized (or migrated externally) instead of in-place evolution

### Requirement: Store raw article HTML as pending work
The repository SHALL store fetched article HTML as latest fetch state keyed by article source and SHALL persist article publication time as article-source metadata when publish metadata is available in fetched HTML. Publication-time parsing SHALL prefer `article:published_time` metadata and fall back to JSON-LD `datePublished` when the primary metadata is absent.

#### Scenario: New raw article
- **WHEN** a new article URL is discovered and fetched
- **THEN** the repository SHALL create an `article_sources` row for the URL
- **AND** it SHALL create one `article_fetches` row for that source
- **AND** it SHALL store the article publication time on the article source when the fetched HTML contains parseable publish metadata

#### Scenario: Duplicate raw article URL
- **WHEN** an article URL already exists and is fetched again
- **THEN** the repository SHALL keep one `article_sources` row for the URL
- **AND** it SHALL upsert the existing `article_fetches` row without creating fetch history
- **AND** it SHALL update the article source publication time when the fetched HTML contains parseable publish metadata

#### Scenario: Existing fetched article backfill
- **WHEN** the repository initializes an existing compatible database that has fetched article HTML but missing article publication time values
- **THEN** it SHALL backfill article source publication time values from parseable publish metadata in existing fetched HTML without changing exported JSON shape

#### Scenario: Backfill fails on missing publish metadata
- **WHEN** an article source has fetched article HTML but no parseable publish metadata
- **THEN** the backfill SHALL fail with an actionable error instead of silently leaving publication time empty for that article source

### Requirement: Query latest fetched articles for downstream processing
The repository SHALL provide access to latest fetched article payloads for downstream processing.

#### Scenario: Latest fetches exist
- **WHEN** `article_fetches` rows exist for discovered sources
- **THEN** the repository SHALL return fetch IDs, source IDs, URLs, and HTML

### Requirement: Export query supports frontend JSON generation
The repository SHALL provide a query that joins geocoded spots, spot links, presenters, and stored article publication time into export-ready data.

#### Scenario: Exporting current map data
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return deduplicated spot records with coordinates and associated presenter names
- **AND** it SHALL make associated article publication time available internally for presenter ordering
- **AND** it SHALL fail with an actionable error if an exportable article has no stored publication time
- **AND** it SHALL NOT add article publication time fields to exported spot or presenter records

### Requirement: Stage failures are surfaced with diagnostics
The repository layer and adapters SHALL surface write/process failures as errors with actionable context.

#### Scenario: Persisting a stage write fails
- **WHEN** a stage cannot persist required data due to constraint or storage errors
- **THEN** the call SHALL return a non-nil error with stage-relevant context
- **AND** callers SHALL treat the stage as failed

### Requirement: Store per-fetch audio as SQLite blobs
The repository SHALL support durable storage of downloaded audio payloads for article-linked videos.

#### Scenario: Insert acquired audio
- **WHEN** audio is acquired for an article fetch-linked video
- **THEN** the repository SHALL store an `audio_sources` row with `article_fetch_id`, `youtube_url`, `audio_format`, `mime_type`, `byte_size`, and `audio_blob`

#### Scenario: Duplicate audio for same fetch
- **WHEN** an `audio_sources` row already exists for an `article_fetch_id`
- **THEN** the repository SHALL upsert the same logical row without creating duplicate ownership records

### Requirement: SQLite schema SHALL store extracted article text content
The repository schema SHALL include dedicated storage for extracted article text linked to latest fetch rows.

#### Scenario: Fresh database initialization includes article text table
- **WHEN** the repository initializes schema on a fresh database
- **THEN** it SHALL create `article_texts` linked to `article_fetch_id`

#### Scenario: Existing database initialization remains idempotent
- **WHEN** the repository initializes schema on an existing database
- **THEN** it SHALL create any missing article text schema objects without dropping existing data
- **AND** repeated initialization SHALL remain idempotent

### Requirement: Repository writes SHALL preserve one authoritative text row per fetch
The repository SHALL preserve one authoritative latest text extraction output per fetch via upsert semantics.

#### Scenario: Persist successful extraction
- **WHEN** extraction content is saved for an article fetch
- **THEN** the repository SHALL upsert one `article_texts` row for that `article_fetch_id`
- **AND** subsequent saves for the same fetch SHALL replace prior text content

### Requirement: Stage-mode support matrix is enforced by command-layer validation
The system SHALL enforce declared stage/mode support and reject unsupported combinations before processing.

#### Scenario: Unsupported stage/mode requested
- **WHEN** a stage is requested with an unsupported I/O mode
- **THEN** validation SHALL fail before any data mutation
- **AND** the command SHALL return a non-zero error with guidance

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

