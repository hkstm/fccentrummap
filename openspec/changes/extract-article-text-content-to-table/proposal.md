## Why

We currently keep full raw HTML for scraped FC Centrum articles, but we do not store the cleaned article text in a queryable form. Extracting and persisting article text now enables downstream processing to work on canonical content instead of repeatedly parsing HTML.

## What Changes

- Add a new extraction capability that parses article HTML and collects text content from article body elements.
- Use Trafilatura to extract cleaned main article text from stored raw HTML in a single robust path.
- Add a dedicated SQLite table to persist extracted text content linked to `article_raw_id`.
- Add repository-layer methods to insert and read extracted article text rows.
- Integrate extraction + persistence into the scraper pipeline after raw HTML fetch succeeds.

## Capabilities

### New Capabilities
- `article-text-content-extraction`: Extract normalized main-content text segments via Trafilatura and persist them as structured rows linked to raw articles.

### Modified Capabilities
- `web-scraper`: Extend post-fetch behavior to run article text extraction from stored HTML and persist extracted segments.
- `sqlite-storage`: Extend schema/repository requirements with a new table and CRUD behavior for extracted article text content.

## Impact

- Affected code: scraper pipeline flow, HTML parsing logic, repository schema and methods.
- Affected data model: new SQLite table for article text segments keyed by `article_raw_id`.
- APIs/dependencies: no external API changes; existing Go stack and SQLite driver remain sufficient.
- Operational impact: incremental DB growth from extracted text rows; improved downstream queryability and reuse.
