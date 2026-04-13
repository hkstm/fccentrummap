## Why

FC Centrum publishes "De Spots van" articles with useful source material, but the site offers no structured way to collect or persist those pages for downstream processing. We needed a durable scraper/storage foundation so later extraction work could operate on a local SQLite-backed corpus instead of scraping ad hoc every time.

## What Changes

- Build a Go scraper that crawls the paginated `fccentrum.nl/categorie/spots/` pages
- Fetch article HTML and store it in SQLite as pending raw-ingestion work
- Establish the normalized SQLite schema and repository layer used by later pipeline stages
- Provide export-ready repository support that can feed downstream frontend JSON generation

## Capabilities

### New Capabilities
- `web-scraper`: Crawl the spots category pages, discover article URLs, and store raw article HTML for later processing
- `sqlite-storage`: SQLite schema and repository/data-access support for raw article storage, pending-work queries, normalized spot relations, and export-ready reads

### Modified Capabilities
None

## Impact

- New Go module under `scraper/` using `colly` and `modernc.org/sqlite`
- Creates and reads `data/spots.db` for the scraper pipeline
- Establishes the storage foundation used by later export and transcript-first extraction work
