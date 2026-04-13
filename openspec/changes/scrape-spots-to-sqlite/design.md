## Context

FC Centrum (fccentrum.nl) publishes "De Spots van" articles where Amsterdam locals share their favorite places. The site is a WordPress/Elementor site with 4 paginated category pages (~12 articles each). Each article contains a YouTube video and a `<figcaption>` that lists spots in a consistent format: `Spot N: Name, Address`. We need to extract this into a structured SQLite database with geocoded coordinates.

There is no existing codebase — this is a greenfield Go project.

## Goals / Non-Goals

**Goals:**
- Reliably scrape all spot articles across all paginated category pages
- Extract author names and spot details (name, address) from article HTML
- Geocode spot addresses to lat/lng using Google Maps Geocoding API
- Store all data in a normalized SQLite database
- Make the scraper re-runnable (idempotent — skip already-scraped articles)

**Non-Goals:**
- Building the Google Maps frontend (future work)
- Scraping non-"spots" categories from fccentrum.nl
- Real-time or scheduled scraping — this is a one-off batch tool
- Automatically fixing articles that don't follow the standard "Spot N: Name, Address" format (but they are logged for manual review)

## Decisions

### 1. Single CLI binary with separated concerns
**Choice:** One binary with `main.go` as the driver orchestrating distinct packages/files under the Go module rooted at `scraper/`:
```
scraper/cmd/scraper/main.go  # CLI entry point, orchestrates the pipeline
scraper/internal/scraper/    # Colly-based HTML fetching, stores raw HTML
scraper/internal/extractor/  # Gemini-based structured data extraction from HTML
scraper/internal/geocoder/   # Google Maps Geocoding API interaction
scraper/internal/repository/ # SQLite database access (read/write/schema)
scraper/internal/models/     # Shared data types (Author, Article, Spot, etc.)
```
**Rationale:** The dataset is small (~40-50 articles, ~150 spots) so a single binary is fine, but separating scraping, geocoding, and DB logic into their own files keeps things testable and readable. `main.go` just wires the pieces together.
**Alternatives considered:** Separate CLI commands for scrape/geocode/store — rejected as over-engineering for this dataset size.

### 2. Web scraping with Colly
**Choice:** Use `github.com/gocolly/colly` for fetching HTML pages.
**Rationale:** Colly is a full scraping framework with built-in request scheduling, rate limiting, and callback-based HTML traversal. Good opportunity to gain familiarity for future projects. The site renders HTML server-side, so no headless browser is needed.
**Alternatives considered:** `goquery` directly — lighter but misses out on Colly's built-in rate limiting and request management.

### 2b. LLM-based extraction with Gemini
**Choice:** Use Google Gemini (via `google.golang.org/genai`) with structured JSON output to extract author names and spot details from raw article HTML.
**Rationale:** The article HTML has inconsistent formatting that makes regex parsing brittle. Gemini's structured output mode guarantees a JSON response matching our schema. The free tier easily covers ~43 articles. Requires `GEMINI_API_KEY` env var.
**Alternatives considered:** Regex parsing — faster and free but broke on real data due to HTML formatting variations.

### 3. Pure Go SQLite with modernc.org/sqlite
**Choice:** Use `modernc.org/sqlite` via `database/sql`.
**Rationale:** Pure Go implementation, no CGo dependency. Simpler builds and cross-compilation. Already proven in user's other projects.
**Alternatives considered:** `mattn/go-sqlite3` — requires CGo toolchain.

### 4. Google Maps Geocoding API via googlemaps Go client
**Choice:** Use `googlemaps/google-maps-services-go` for geocoding.
**Rationale:** Official Google client library. Handles rate limiting and retries. All spot addresses are in Amsterdam, so we append ", Amsterdam" to improve accuracy.
**Alternatives considered:** Raw HTTP calls to the Geocoding API — more boilerplate with no benefit.

### Fail-fast on geocoding errors
**Choice:** If any geocoding request fails, abort the entire run immediately — do not attempt to geocode remaining spots.
**Rationale:** Protects against burning through free tier API quota when something is systematically wrong (bad API key, quota exhausted, malformed addresses). Already-processed articles remain `COMPLETED`; the current article stays `PENDING` so it can be retried on the next run.

### 5. Database schema

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

**Rationale:** The pipeline is split into stages via `articles_raw` as an intermediate table:
1. **Scrape** — fetch HTML, store in `articles_raw` with status `PENDING`
2. **Parse + Geocode** — extract author/spots, geocode addresses. On success: insert into `authors`, `articles`, `spots`, `article_spots` and set status to `COMPLETED`. On failure: set status to `FAILED`.

This means `articles_raw` is the source of truth for what's been scraped, and re-runs can pick up `PENDING` or retry `FAILED` entries. The `spots` table only contains fully geocoded entries (non-nullable lat/lng), keeping it always clean and ready for map display. `articles` links back to `articles_raw` via `article_raw_id` for traceability.

### 6. Scraping strategy
**Choice:** Two-phase approach:
1. Crawl category pages (`/categorie/spots/page/N/`) to collect all article URLs
2. Fetch each article page and parse the `<figcaption>` content for spot listings

**Rationale:** The category pages use `data-max-page="4"` to indicate total pages. Article links are in `<a>` tags wrapping each loop item. The figcaption contains spot data in the pattern `Spot N: Name, Address`.

## Risks / Trade-offs

- **[HTML structure changes]** → The scraper depends on Elementor's current markup. If fccentrum.nl redesigns, parsing will break. Mitigation: log warnings for articles that don't match expected patterns; the dataset is small enough to manually verify.
- **[Inconsistent spot format]** → Some articles may not list spots in the `Spot N: Name, Address` format. Mitigation: mark as `FAILED` in `articles_raw` with an error message. Raw HTML is preserved for manual review or future re-parsing.
- **[Geocoding costs]** → Google Maps Geocoding API charges per request. Mitigation: fail-fast on any geocoding error to avoid wasting quota. Idempotent design means re-runs skip already-completed articles.
- **[Rate limiting by fccentrum.nl]** → Mitigation: add a small delay between HTTP requests (e.g., 500ms) to be respectful.
