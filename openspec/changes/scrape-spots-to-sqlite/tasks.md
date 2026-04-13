## 1. Project Setup

- [x] 1.1 Initialize Go module (`go mod init`) and create directory structure: `cmd/scraper/`, `internal/models/`, `internal/scraper/`, `internal/geocoder/`, `internal/repository/`
- [x] 1.2 Add dependencies: `github.com/gocolly/colly`, `modernc.org/sqlite`, `googlemaps/google-maps-services-go`

## 2. Models

- [x] 2.1 Define Go structs in `internal/models/`: `ArticleRaw`, `Author`, `Spot`, `Article`, `ArticleSpot`

## 3. Repository (SQLite)

- [x] 3.1 Implement schema initialization — create all tables (`articles_raw`, `authors`, `spots`, `articles`, `article_spots`) if they don't exist
- [x] 3.2 Implement `InsertArticleRaw` — insert URL + HTML with status `PENDING`, skip on duplicate URL
- [x] 3.3 Implement `GetPendingArticles` — query all `articles_raw` entries with status `PENDING`
- [x] 3.4 Implement `InsertAuthor` — insert or get existing author by name, return `author_id`
- [x] 3.5 Implement `InsertSpot` — insert or get existing spot by name + address, return `spot_id`
- [x] 3.6 Implement `InsertArticle` — insert into `articles` table linking `article_raw_id` and `author_id`
- [x] 3.7 Implement `LinkArticleSpots` — insert rows into `article_spots`
- [x] 3.8 Implement `UpdateArticleRawStatus` — set status to `COMPLETED` or `FAILED`, refresh `updated_at`, log error with `article_raw_id`, URL, and reason on failure

## 4. Scraper

- [x] 4.1 Implement category page crawler — use Colly to fetch paginated `/categorie/spots/page/N/` pages, read `data-max-page`, extract article URLs from loop item `<a>` elements, deduplicate
- [x] 4.2 Implement article fetcher — use Colly to fetch each article URL and store raw HTML via `InsertArticleRaw`, with 500ms delay between requests
- [x] 4.3 Implement article parser — extract author name from title pattern "DE SPOTS VAN: [NAME]" (title-cased), log error with URL + actual title on mismatch
- [x] 4.4 Implement spot parser — extract spots from `<figcaption>` matching pattern `Spot N: Name, Address`, log error with URL + raw figcaption text on failure

## 5. Geocoder

- [x] 5.1 Implement API key loading from `GOOGLE_MAPS_API_KEY` env var, log and exit if missing
- [x] 5.2 Implement geocoding function — resolve address + ", Amsterdam" to lat/lng using Google Maps client
- [x] 5.3 Implement fail-fast behavior — on any geocoding error, log spot name, address, and API error, then abort the run

## 6. CLI Driver

- [x] 6.1 Implement `main.go` pipeline: init DB → validate API key → crawl category pages → fetch articles → process pending articles (parse → geocode → store) → log summary
- [x] 6.2 Ensure idempotency — skip already-scraped URLs, only process `PENDING` articles, preserve `COMPLETED` entries on abort

## 7. Verification

- [ ] 7.1 Run scraper against live site and verify `articles_raw` is populated with correct HTML and URLs
- [ ] 7.2 Run full pipeline and verify `authors`, `articles`, `spots`, `article_spots` are populated correctly
- [ ] 7.3 Verify re-run is idempotent — no duplicate entries, `PENDING`/`FAILED` articles are retried
- [ ] 7.4 Verify fail-fast geocoding — confirm the run aborts on a bad API key and no spots are inserted
