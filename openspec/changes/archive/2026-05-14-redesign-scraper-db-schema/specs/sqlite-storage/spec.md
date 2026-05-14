## MODIFIED Requirements

### Requirement: Initialize schema on startup
The scraper repository layer SHALL create required tables if they do not already exist.

#### Scenario: Fresh database
- **WHEN** the program opens a new `data/spots.db`
- **THEN** it SHALL create `article_sources`, `article_fetches`, `article_texts`, `audio_sources`, `audio_transcriptions`, `spot_mentions`, `spot_google_geocodes`, `presenters`, `article_presenters`, and `article_spots`
- **AND** it SHALL enforce uniqueness for transcription rows on `(audio_source_id, provider, language)`
- **AND** it SHALL enforce one-row-per-source latest fetch uniqueness on `article_fetches(article_source_id)`

#### Scenario: Existing database
- **WHEN** the program opens an existing database with the required schema
- **THEN** existing data SHALL be preserved
- **AND** missing newly required schema objects SHALL be created idempotently

### Requirement: Export query supports frontend JSON generation
The repository SHALL provide a query that joins geocoded spots, spot links, and presenters into export-ready data.

#### Scenario: Exporting current map data
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return deduplicated spot records with coordinates and associated presenter names

### Requirement: Store raw article HTML as pending work
The repository SHALL store fetched article HTML as latest fetch state keyed by article source.

#### Scenario: New raw article
- **WHEN** a new article URL is discovered and fetched
- **THEN** the repository SHALL create an `article_sources` row for the URL
- **AND** it SHALL create one `article_fetches` row for that source

#### Scenario: Duplicate raw article URL
- **WHEN** an article URL already exists and is fetched again
- **THEN** the repository SHALL keep one `article_sources` row for the URL
- **AND** it SHALL upsert the existing `article_fetches` row without creating fetch history

## REMOVED Requirements

### Requirement: Current change defers new geocode-to-final-table SQLite writes
**Reason**: v2 schema now includes direct SQLite writes for geocoding and final spot linking.
**Migration**: Replace file-handoff assumptions with SQLite-backed `geocode-spots` writes to `spot_google_geocodes` and `article_spots`.
