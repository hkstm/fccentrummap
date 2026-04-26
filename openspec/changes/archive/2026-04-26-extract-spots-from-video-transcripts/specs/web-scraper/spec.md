## MODIFIED Requirements

### Requirement: Fetch article HTML and store it as raw input
The scraper SHALL fetch article HTML, store it as raw input, and detect embedded YouTube videos for downstream audio acquisition.

#### Scenario: Successful fetch with embedded video
- **WHEN** an article page responds successfully and contains an embedded YouTube video
- **THEN** the scraper SHALL store the raw HTML
- **AND** it SHALL persist a nullable string field `video_id` on the raw article record
- **AND** `video_id` SHALL contain the normalized canonical YouTube ID (11-character ID only, no URL prefix)

#### Scenario: Successful fetch without embedded video
- **WHEN** an article page responds successfully and does not contain an embedded YouTube video
- **THEN** the scraper SHALL still store the raw HTML without failing the fetch phase

#### Scenario: Respectful request pacing
- **WHEN** the scraper makes requests to `fccentrum.nl`
- **THEN** it SHALL apply a delay between requests instead of issuing them as a tight burst

### Requirement: Current scraper run stops after raw-ingestion stage
The current scraper CLI SHALL continue into YouTube-audio acquisition after crawling and raw HTML storage.

#### Scenario: End of current pipeline
- **WHEN** `scraper/cmd/scraper` finishes a run
- **THEN** it SHALL attempt audio acquisition for rows with detected `video_id`
- **AND** it SHALL persist acquired audio blobs into `article_audio_sources`
