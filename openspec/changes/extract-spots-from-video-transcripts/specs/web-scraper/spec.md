## MODIFIED Requirements

### Requirement: Discover article URLs from paginated spots category pages
The scraper SHALL crawl `https://fccentrum.nl/categorie/spots/` and discover article URLs from paginated category pages.

#### Scenario: Crawl category pages
- **WHEN** the scraper runs the article discovery phase
- **THEN** it SHALL inspect the base category page, read the available pagination range, and visit each category page in turn

#### Scenario: No duplicate URLs
- **WHEN** an article link appears more than once during discovery
- **THEN** the scraper SHALL keep only one copy of that URL in the crawl result

### Requirement: Fetch article HTML and detect embedded YouTube videos
The scraper SHALL fetch article HTML, store it as raw input, and detect embedded YouTube videos needed for transcript-based extraction.

#### Scenario: Successful fetch with embedded video
- **WHEN** an article page responds successfully and contains an embedded YouTube video
- **THEN** the scraper SHALL store the raw HTML and make the embedded video identifier available for downstream transcript extraction

#### Scenario: Successful fetch without embedded video
- **WHEN** an article page responds successfully and does not contain an embedded YouTube video
- **THEN** the scraper SHALL still store the raw HTML without failing the fetch phase
