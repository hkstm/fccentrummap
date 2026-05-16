## MODIFIED Requirements

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

### Requirement: Export query supports frontend JSON generation
The repository SHALL provide a query that joins geocoded spots, spot links, presenters, and stored article publication time into export-ready data.

#### Scenario: Exporting current map data
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return deduplicated spot records with coordinates and associated presenter names
- **AND** it SHALL make associated article publication time available internally for presenter ordering
- **AND** it SHALL fail with an actionable error if an exportable article has no stored publication time
- **AND** it SHALL NOT add article publication time fields to exported spot or presenter records
