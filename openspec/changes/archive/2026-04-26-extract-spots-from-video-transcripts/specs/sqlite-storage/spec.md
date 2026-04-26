## MODIFIED Requirements

### Requirement: Initialize schema on startup
The scraper repository layer SHALL create required tables if they do not already exist.

#### Scenario: Fresh database
- **WHEN** the program opens a new `data/spots.db`
- **THEN** it SHALL create `articles_raw`, `authors`, `spots`, `articles`, `article_spots`, and `article_audio_sources`

#### Scenario: Existing database
- **WHEN** the program opens an existing database with the required schema
- **THEN** existing data SHALL be preserved

## ADDED Requirements

### Requirement: Store per-article audio as SQLite blobs
The repository SHALL support durable storage of downloaded audio payloads for article-linked videos.

#### Scenario: Insert acquired audio
- **WHEN** audio is acquired for an article-linked video
- **THEN** the repository SHALL store an `article_audio_sources` row with `article_raw_id`, `video_id`, `youtube_url`, `audio_format`, `mime_type`, `byte_size`, and `audio_blob`

#### Scenario: Duplicate audio for same article
- **WHEN** an `article_audio_sources` row already exists for an `article_raw_id`
- **THEN** the pipeline SHALL skip duplicate insertion unless explicitly forced
