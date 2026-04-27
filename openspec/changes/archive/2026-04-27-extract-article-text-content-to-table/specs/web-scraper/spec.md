## ADDED Requirements

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
