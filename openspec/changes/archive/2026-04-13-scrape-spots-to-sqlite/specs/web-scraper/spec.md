## ADDED Requirements

### Requirement: Discover all article URLs from paginated category pages
The scraper SHALL crawl `https://fccentrum.nl/categorie/spots/` and all subsequent paginated pages to collect every article URL. It SHALL read the `data-max-page` attribute to determine the total number of pages.

#### Scenario: Scrape all pages
- **WHEN** the scraper starts
- **THEN** it SHALL fetch pages 1 through N (where N is `data-max-page`) and extract all article URLs from the loop item `<a>` elements

#### Scenario: No duplicate URLs
- **WHEN** the same article URL appears on multiple pages
- **THEN** it SHALL be collected only once

### Requirement: Store raw article HTML in articles_raw
The scraper SHALL fetch each discovered article URL and store the full HTML response body in the `articles_raw` table with status `PENDING`.

#### Scenario: New article fetched successfully
- **WHEN** an article URL is fetched and its URL does not exist in `articles_raw`
- **THEN** the HTML SHALL be inserted into `articles_raw` with status `PENDING`

#### Scenario: Article already exists in articles_raw
- **WHEN** an article URL already exists in `articles_raw`
- **THEN** the scraper SHALL skip duplicate insertion/storage for that URL

### Requirement: Parsing and structured extraction are deferred
Author/spot parsing from article content is intentionally deferred in this archived change and handled by successor transcript-first extraction work.

#### Scenario: Deferred extraction scope
- **WHEN** contributors read this archived change
- **THEN** they SHALL treat author/spot parsing as out of scope for this change
- **AND** they SHALL use `extract-spots-from-video-transcripts` as the successor change for extraction behavior

### Requirement: Respect rate limiting
The scraper SHALL add a delay between HTTP requests to fccentrum.nl to avoid overwhelming the server.

#### Scenario: Delay between requests
- **WHEN** the scraper fetches consecutive pages or articles
- **THEN** it SHALL wait at least 500ms between requests
