## MODIFIED Requirements

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
