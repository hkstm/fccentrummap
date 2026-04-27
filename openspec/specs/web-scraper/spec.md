## Purpose

Define the canonical raw-ingestion behavior of the current FC Centrum scraper pipeline.
## Requirements
### Requirement: Discover article URLs from paginated spots category pages
The scraper SHALL crawl `https://fccentrum.nl/categorie/spots/` and discover article URLs from paginated category pages.

#### Scenario: Crawl category pages
- **WHEN** the scraper runs the article discovery phase
- **THEN** it SHALL inspect the base category page, read the available pagination range, and visit each category page in turn

#### Scenario: No duplicate URLs
- **WHEN** an article link appears more than once during discovery
- **THEN** the scraper SHALL keep only one copy of that URL in the crawl result

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

### Requirement: Current scraper run continues into YouTube audio acquisition
The current scraper CLI SHALL continue into YouTube-audio acquisition after crawling and raw HTML storage.

#### Scenario: End of current pipeline
- **WHEN** `scraper/cmd/scraper` finishes a run
- **THEN** it SHALL attempt audio acquisition for rows with detected `video_id`
- **AND** it SHALL persist acquired audio blobs into `article_audio_sources`

### Requirement: Scraper SHALL run article text extraction after raw HTML is stored
The scraper pipeline SHALL execute article text extraction for each successfully fetched article and persist extraction outcomes as part of normal processing.

#### Scenario: Successful fetch with usable extraction
- **WHEN** an article page is fetched and Trafilatura yields usable article text
- **THEN** the scraper SHALL persist a successful extraction outcome for that article
- **AND** it SHALL persist extracted text segments associated with that outcome

#### Scenario: Successful fetch with insufficient extracted text
- **WHEN** an article page is fetched but extraction yields insufficient/empty content
- **THEN** the scraper SHALL persist a `no_match` extraction outcome for that article
- **AND** it SHALL continue processing remaining pipeline steps without failing the scraper run

#### Scenario: Extraction runtime error
- **WHEN** an article page is fetched but extraction fails due to parser/runtime error
- **THEN** the scraper SHALL persist an `error` extraction outcome for that article
- **AND** it SHALL treat this as extraction-only failure while continuing the overall scraper run

