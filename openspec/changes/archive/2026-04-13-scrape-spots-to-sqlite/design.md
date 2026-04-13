## Context

FC Centrum (fccentrum.nl) publishes "De Spots van" articles where Amsterdam locals share their favorite places. The site is a WordPress/Elementor site with paginated category pages and article pages that need to be captured locally before later extraction stages can operate on them.

This change ended up establishing the scraper/storage foundation rather than the full end-to-end extraction pipeline originally imagined. The current implemented behavior is: discover article URLs, fetch/store raw HTML, initialize the SQLite schema, and report pending raw articles for the next transcript-first extraction stage.

## Goals / Non-Goals

**Goals:**
- Reliably scrape all spot article URLs across paginated category pages
- Store raw article HTML in SQLite as pending downstream work
- Establish the normalized SQLite schema and repository layer used by later stages
- Make the scraper re-runnable (idempotent — skip already-ingested URLs)
- Provide the storage/query foundation for downstream export and transcript-first extraction work

**Non-Goals:**
- Completing article-text extraction in the current pipeline
- Completing transcript extraction in this change
- Running a full parse -> geocode -> normalized ingest loop as part of the current scraper CLI
- Building the Google Maps frontend

## Decisions

### 1. Single CLI binary with separated concerns
**Choice:** One binary with `main.go` as the driver orchestrating distinct packages/files under the Go module rooted at `scraper/`:
```
scraper/cmd/scraper/main.go  # CLI entry point for crawl/fetch/raw-ingestion flow
scraper/internal/scraper/    # Colly-based discovery and HTML fetching
scraper/internal/repository/ # SQLite database access (read/write/schema/export query)
scraper/internal/models/     # Shared data types
scraper/internal/geocoder/   # Early geocoding groundwork retained as library code
```
**Rationale:** The dataset is small enough for a single binary, but separating scraping and repository logic keeps the foundation reusable for later extraction stages.
**Alternatives considered:** Separate CLI commands for crawl/store phases — rejected as unnecessary at this stage.

### 2. Web scraping with Colly
**Choice:** Use `github.com/gocolly/colly` for fetching HTML pages.
**Rationale:** Colly is a full scraping framework with built-in request scheduling, rate limiting, and callback-based HTML traversal. Good opportunity to gain familiarity for future projects. The site renders HTML server-side, so no headless browser is needed.
**Alternatives considered:** `goquery` directly — lighter but misses out on Colly's built-in rate limiting and request management.

### 3. Pure Go SQLite with modernc.org/sqlite
**Choice:** Use `modernc.org/sqlite` via `database/sql`.
**Rationale:** Pure Go implementation, no CGo dependency. Simpler builds and cross-compilation. Already proven in user's other projects.
**Alternatives considered:** `mattn/go-sqlite3` — requires CGo toolchain.

### 4. Database schema

```sql
CREATE TABLE articles_raw (
    article_raw_id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    html TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',  -- PENDING, COMPLETED, FAILED
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE authors (
    author_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE spots (
    spot_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    UNIQUE(name, address)
);

CREATE TABLE articles (
    article_id INTEGER PRIMARY KEY AUTOINCREMENT,
    article_raw_id INTEGER NOT NULL REFERENCES articles_raw(article_raw_id),
    author_id INTEGER NOT NULL REFERENCES authors(author_id),
    title TEXT NOT NULL
);

CREATE TABLE article_spots (
    article_id INTEGER NOT NULL REFERENCES articles(article_id),
    spot_id INTEGER NOT NULL REFERENCES spots(spot_id),
    PRIMARY KEY (article_id, spot_id)
);
```

**Rationale:** The foundation is split into stages via `articles_raw` as the durable raw-ingestion table:
1. **Discover** — crawl paginated category pages to collect article URLs
2. **Fetch + Store** — fetch article HTML, store it in `articles_raw` with status `PENDING`
3. **Downstream processing (later changes)** — transcript-first extraction and later persistence updates consume pending rows

This means `articles_raw` is the source of truth for what has been scraped locally, and re-runs can skip duplicate URLs while later pipeline stages consume pending rows.

### 5. Scraping strategy
**Choice:** Two-phase approach:
1. Crawl category pages (`/categorie/spots/page/N/`) to collect all article URLs
2. Fetch each article page and store the raw HTML in SQLite

**Rationale:** The category pages expose pagination via `data-max-page`, and article links can be collected from the loop-item anchor structure. Storing raw HTML first keeps later extraction logic decoupled from network fetches.

## Risks / Trade-offs

- **[HTML structure changes]** → The scraper depends on current site markup. Mitigation: keep discovery/fetch logic small and easy to update.
- **[Raw corpus without full extraction yet]** → The foundation alone does not produce a fully normalized spot dataset. Mitigation: later changes (`extract-spots-from-video-transcripts`) are the intended successor for extraction behavior.
- **[Rate limiting by fccentrum.nl]** → Mitigation: add a small delay between HTTP requests (e.g., 500ms) to be respectful.
