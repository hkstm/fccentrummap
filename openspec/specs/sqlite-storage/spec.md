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

