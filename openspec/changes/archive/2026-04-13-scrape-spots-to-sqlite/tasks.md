## 1. Project setup

- [x] 1.1 Initialize the Go module under `scraper/` and create the scraper/repository/model package structure
- [x] 1.2 Add core dependencies for crawling and SQLite storage

## 2. Storage foundation

- [x] 2.1 Implement schema initialization for `articles_raw`, `authors`, `spots`, `articles`, and `article_spots`
- [x] 2.2 Implement raw-HTML insertion with duplicate URL protection
- [x] 2.3 Implement pending-article queries for downstream processing
- [x] 2.4 Implement normalized repository helpers and failure-status logging used by later pipeline stages
- [x] 2.5 Implement export-ready repository query support for downstream JSON generation

## 3. Scraper foundation

- [x] 3.1 Implement paginated category crawling and URL discovery
- [x] 3.2 Implement article fetch-and-store flow with respectful request pacing
- [x] 3.3 Make the current scraper CLI stop after raw-ingestion and pending-count reporting
- [x] 3.4 Keep the raw-ingestion flow idempotent for repeated runs

## 4. Verification

- [x] 4.1 Verify the scraper module builds successfully
- [x] 4.2 Verify the repository and export foundation support downstream pipeline work
- [x] 4.3 Confirm later extraction work will build on pending raw-article storage rather than re-fetching ad hoc
