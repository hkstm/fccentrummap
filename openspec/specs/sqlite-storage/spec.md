## Purpose

Define the canonical SQLite schema and repository-layer behavior used by the current scraper and exporter.
## Requirements
### Requirement: Initialize schema on startup
The scraper repository layer SHALL create required tables and compatible schema extensions if they do not already exist, including Gemini-direct source tables used by source-selected downstream stages.

#### Scenario: Fresh database
- **WHEN** the program opens a new `data/spots.db`
- **THEN** it SHALL create `article_sources`, `article_fetches`, `article_texts`, `audio_sources`, `audio_transcriptions`, `spot_mentions`, `spot_google_geocodes`, `presenters`, `article_presenters`, and `article_spots`
- **AND** it SHALL create `gemini_direct_spot_mentions`, `gemini_direct_spot_google_geocodes`, `gemini_direct_article_spots`, and `gemini_direct_spot_corrections`
- **AND** `article_sources` SHALL include nullable publication timestamp storage for article publish time
- **AND** it SHALL enforce uniqueness for transcription rows on `(audio_source_id, provider, language)`
- **AND** it SHALL enforce one-row-per-source latest fetch uniqueness on `article_fetches(article_source_id)`
- **AND** it SHALL enforce one-row-per-source Gemini mention uniqueness on `gemini_direct_spot_mentions(article_source_id, place)`

#### Scenario: Existing database
- **WHEN** the program opens an existing database with the required schema
- **THEN** existing data SHALL be preserved
- **AND** missing newly required schema objects or compatible columns, including article publication timestamp storage and Gemini-direct source tables, SHALL be created idempotently for compatible v2-era databases
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

### Requirement: SQLite schema stores spot corrections
The repository schema SHALL include durable storage for manual spot corrections keyed by stable spot identifier.

#### Scenario: Fresh database initialization includes corrections table
- **WHEN** the repository initializes a fresh SQLite database
- **THEN** it SHALL create a `spot_corrections` table or equivalent correction storage with stable spot identifier, nullable spot name override, nullable place ID override, nullable corrected latitude/longitude, nullable timestamp override, hidden flag, and audit timestamps

#### Scenario: Existing database initialization adds corrections table
- **WHEN** the repository initializes an existing compatible database without correction storage
- **THEN** it SHALL create the missing correction storage idempotently without dropping existing source data

### Requirement: Repository upserts spot corrections
The repository SHALL expose operations to save and retrieve the current correction for a stable spot identifier.

#### Scenario: Insert correction
- **WHEN** no correction exists for a stable spot identifier
- **THEN** saving a correction SHALL insert one correction row for that identifier

#### Scenario: Update correction
- **WHEN** a correction already exists for a stable spot identifier
- **THEN** saving a correction SHALL update that existing correction row
- **AND** it SHALL preserve one active correction per stable spot identifier

### Requirement: Export query includes correction data
The repository SHALL make stored spot corrections available to export assembly without mutating source spot records.

#### Scenario: Exporting corrected and uncorrected spots
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return each exportable spot with its stable spot identifier and any stored correction values
- **AND** uncorrected spots SHALL remain exportable with their original source values

### Requirement: SQLite schema stores Gemini-derived spot mentions
The repository schema SHALL store Gemini-derived spot mentions in a source-specific table that links directly to article source and YouTube URL context.

#### Scenario: Fresh database includes Gemini mention table
- **WHEN** the repository initializes schema on a fresh database
- **THEN** it SHALL create `gemini_direct_spot_mentions` with a primary mention identifier, `article_source_id`, `youtube_url`, `place`, nullable `youtube_timestamp_seconds`, nullable evidence, nullable confidence, nullable model, and extraction timestamp storage
- **AND** `article_source_id` SHALL reference `article_sources(article_source_id)` with cascade deletion
- **AND** `youtube_timestamp_seconds` SHALL be null or non-negative
- **AND** `confidence` SHALL be null or between 0 and 1 inclusive

#### Scenario: Duplicate Gemini mention for an article and place
- **WHEN** the repository is asked to persist a Gemini-derived mention for an article source and place that already exists
- **THEN** it SHALL preserve one logical row for that article source and place
- **AND** it SHALL update source metadata using idempotent upsert semantics instead of inserting a duplicate row

### Requirement: SQLite schema stores Gemini-derived geocodes and article links
The repository schema SHALL store Gemini-derived geocoding results and article-to-spot links in source-specific tables compatible with existing downstream export assembly.

#### Scenario: Fresh database includes Gemini geocode and link tables
- **WHEN** the repository initializes schema on a fresh database
- **THEN** it SHALL create `gemini_direct_spot_google_geocodes` linked to `gemini_direct_spot_mentions`
- **AND** it SHALL create `gemini_direct_article_spots` linking `article_sources` to Gemini geocode rows
- **AND** each Gemini mention SHALL have at most one latest Gemini geocode row
- **AND** foreign-key constraints SHALL cascade deletes from article sources and Gemini mentions to dependent Gemini rows

#### Scenario: Gemini geocode is persisted
- **WHEN** a Gemini-derived spot mention is geocoded successfully
- **THEN** the repository SHALL upsert the latest geocode row for that Gemini mention in `gemini_direct_spot_google_geocodes`
- **AND** it SHALL link the article source to that geocode row in `gemini_direct_article_spots`
- **AND** it SHALL NOT write transcript-derived `spot_google_geocodes` or `article_spots` rows for the same operation

### Requirement: SQLite schema stores Gemini-specific spot corrections
The repository schema SHALL include durable correction storage for Gemini-derived exported spot identifiers separate from transcript-derived correction storage.

#### Scenario: Fresh database includes Gemini correction table
- **WHEN** the repository initializes schema on a fresh database
- **THEN** it SHALL create `gemini_direct_spot_corrections` with stable spot identifier, nullable spot name override, nullable place ID override, nullable corrected latitude/longitude, nullable timestamp override, hidden flag, and audit timestamps
- **AND** it SHALL enforce the same corrected-place coordinate consistency constraints as transcript correction storage

#### Scenario: Gemini correction is saved
- **WHEN** a correction is saved for a Gemini-derived exported spot identifier
- **THEN** the repository SHALL upsert that correction in `gemini_direct_spot_corrections`
- **AND** it SHALL NOT mutate transcript-derived `spot_corrections`

### Requirement: Repository exposes source-neutral spot operations
The repository layer SHALL expose source-neutral operations for downstream stages while routing SQL to transcript-specific or Gemini-specific tables based on the selected spot source.

#### Scenario: Listing ungeocoded mentions by source
- **WHEN** the geocode stage requests ungeocoded spot mentions for `gemini-direct`
- **THEN** the repository SHALL return rows from `gemini_direct_spot_mentions` in a source-neutral DTO shape
- **AND** the DTO SHALL include mention id, article source id, place name, YouTube URL, timestamp, and any available provenance required by geocoding

#### Scenario: Exporting data by source
- **WHEN** the export stage requests export rows for `transcript`
- **THEN** the repository SHALL query transcript-derived spot, geocode, article-link, presenter, and correction tables
- **AND** it SHALL return the same source-neutral export DTO shape used for Gemini-direct exports
