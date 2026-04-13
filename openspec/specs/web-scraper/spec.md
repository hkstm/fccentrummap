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
The scraper SHALL fetch each discovered article page and store the raw HTML in the repository database.

#### Scenario: Successful fetch
- **WHEN** an article page responds successfully
- **THEN** the scraper SHALL insert the article URL and HTML into `articles_raw`

#### Scenario: Respectful request pacing
- **WHEN** the scraper makes requests to `fccentrum.nl`
- **THEN** it SHALL apply a delay between requests instead of issuing them as a tight burst

### Requirement: Current scraper run stops after raw-ingestion stage
The current scraper CLI SHALL stop after crawling, raw HTML storage, and pending-count reporting.

#### Scenario: End of current pipeline
- **WHEN** `scraper/cmd/scraper` finishes a run
- **THEN** it SHALL report the count of pending articles remaining for the future video/transcript extraction stage
- **AND** it SHALL not attempt article-text LLM extraction as part of the current pipeline
