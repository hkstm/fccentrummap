## MODIFIED Requirements

### Requirement: Scraper SQLite schema SHALL be rebuilt as Gemini-first relational schema
The system SHALL initialize a SQLite schema with normalized relational tables for article sources, article fetches, presenters, Gemini-direct spot mentions, Gemini-direct geocodes, Gemini-direct article links, and Gemini-direct spot corrections. The schema SHALL NOT create legacy transcription-path extraction tables.

#### Scenario: Fresh initialization
- **WHEN** `scrape init` is run against a fresh database
- **THEN** `article_sources`, `article_fetches`, `presenters`, `article_presenters`, `gemini_direct_spot_mentions`, `gemini_direct_spot_google_geocodes`, `gemini_direct_article_spots`, and `gemini_direct_spot_corrections` SHALL be created
- **AND** foreign-key and uniqueness constraints SHALL be active
- **AND** `gemini_direct_spot_mentions` SHALL include nullable `address` text storage
- **AND** `gemini_direct_spot_google_geocodes` SHALL include nullable `primary_type_display_name` storage
- **AND** legacy tables `article_texts`, `audio_sources`, `audio_transcriptions`, `spot_mentions`, `spot_google_geocodes`, `article_spots`, and `spot_corrections` SHALL NOT be required by the current pipeline schema

### Requirement: Article fetch storage SHALL be latest-only per source
The schema SHALL store one latest fetch row per `article_source_id` in `article_fetches`.

#### Scenario: Re-fetch existing article source
- **WHEN** `fetch-articles` processes an already-known `article_source_id`
- **THEN** the existing `article_fetches` row SHALL be upserted
- **AND** no additional historical fetch row SHALL be created

### Requirement: Gemini geocode stage SHALL persist only latest successful geocode row per mention
The schema SHALL keep one latest successful geocode row per `gemini_direct_spot_mention_id` in `gemini_direct_spot_google_geocodes` with no retry-history table.

#### Scenario: Geocode Gemini mention multiple times
- **WHEN** `geocode-spots` reruns for a previously geocoded Gemini mention
- **THEN** the single `gemini_direct_spot_google_geocodes` row for that mention SHALL be updated/upserted
- **AND** no separate retry history rows SHALL be created

## REMOVED Requirements

### Requirement: Cleaned article text SHALL be persisted as one row per fetch
**Reason**: Cleaned article text was only required by the legacy transcript-based extraction path. Gemini-direct extraction uses the article URL and YouTube URL as source inputs.
**Migration**: Stop creating and consuming `article_texts`; retain `article_fetches` for article metadata and YouTube URL extraction/fallbacks.
