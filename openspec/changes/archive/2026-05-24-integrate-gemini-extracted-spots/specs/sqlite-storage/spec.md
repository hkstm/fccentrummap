## MODIFIED Requirements

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

## ADDED Requirements

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
