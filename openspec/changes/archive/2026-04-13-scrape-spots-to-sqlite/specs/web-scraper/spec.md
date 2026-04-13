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
- **THEN** the scraper SHALL skip fetching it

### Requirement: Parse author name from article HTML
The scraper SHALL extract the author name from the article title. Article titles follow the pattern "DE SPOTS VAN: [NAME]".

#### Scenario: Standard title format
- **WHEN** an article has a title like "DE SPOTS VAN: NIELS OOSTHOEK"
- **THEN** the extracted author name SHALL be "Niels Oosthoek" (title-cased)

#### Scenario: Non-standard title format
- **WHEN** an article title does not match the expected pattern
- **THEN** the article SHALL be marked as `FAILED` in `articles_raw`
- **AND** the scraper SHALL log an error including the article URL and the actual title that failed to match

### Requirement: Parse spot listings from article HTML
The scraper SHALL extract spot entries from the `<figcaption>` element. Spots follow the pattern `Spot N: Name, Address`.

#### Scenario: Standard spot format
- **WHEN** the figcaption contains "Spot 1: Nationale Opera & Ballet, Amstel 3"
- **THEN** the scraper SHALL extract name "Nationale Opera & Ballet" and address "Amstel 3"

#### Scenario: No spots found
- **WHEN** the figcaption does not contain any entries matching the spot pattern
- **THEN** the article SHALL be marked as `FAILED` in `articles_raw`
- **AND** the scraper SHALL log an error including the article URL and the raw figcaption text that failed to parse

### Requirement: Respect rate limiting
The scraper SHALL add a delay between HTTP requests to fccentrum.nl to avoid overwhelming the server.

#### Scenario: Delay between requests
- **WHEN** the scraper fetches consecutive pages or articles
- **THEN** it SHALL wait at least 500ms between requests
