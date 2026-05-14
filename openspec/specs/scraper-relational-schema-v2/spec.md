# scraper-relational-schema-v2 Specification

## Purpose
TBD - created by archiving change redesign-scraper-db-schema. Update Purpose after archive.
## Requirements
### Requirement: Scraper SQLite schema SHALL be rebuilt as relational v2
The system SHALL initialize a new SQLite schema with normalized relational tables: `article_sources`, `article_fetches`, `article_texts`, `audio_sources`, `audio_transcriptions`, `spot_mentions`, `spot_google_geocodes`, `presenters`, `article_presenters`, and `article_spots`.

#### Scenario: Fresh v2 initialization
- **WHEN** `scrape init` is run against a fresh database
- **THEN** all relational v2 tables SHALL be created
- **AND** foreign-key and uniqueness constraints SHALL be active

### Requirement: Article fetch storage SHALL be latest-only per source
The schema SHALL store one latest fetch row per `article_source_id` in `article_fetches`.

#### Scenario: Re-fetch existing article source
- **WHEN** `fetch-articles` processes an already-known `article_source_id`
- **THEN** the existing `article_fetches` row SHALL be upserted
- **AND** no additional historical fetch row SHALL be created

### Requirement: Cleaned article text SHALL be persisted as one row per fetch
The schema SHALL persist one normalized cleaned text payload per `article_fetch_id` in `article_texts`.

#### Scenario: Extract text for fetched article
- **WHEN** `extract-article-text` succeeds for an `article_fetch_id`
- **THEN** exactly one `article_texts` row SHALL exist for that fetch
- **AND** `cleaned_text` SHALL be stored as non-empty normalized text

### Requirement: Geocode stage SHALL persist only latest successful geocode row per mention
The schema SHALL keep one latest successful geocode row per `spot_mention_id` in `spot_google_geocodes` with no retry-history table.

#### Scenario: Geocode mention multiple times
- **WHEN** `geocode-spots` reruns for a previously geocoded mention
- **THEN** the single `spot_google_geocodes` row for that mention SHALL be updated/upserted
- **AND** no separate retry history rows SHALL be created

