# Scraper SQLite v2 migration note

This project now uses a **breaking** relational v2 schema for scraper persistence.

## Important

Existing SQLite files from the legacy schema are **unsupported**.

Reinitialize using:

```bash
cd scraper
go run ./cmd/scrape init --db-path ../data/spots.db --reset
```

## Canonical stage → table writer ownership

- `collect-article-urls` → `article_sources`
- `fetch-articles` → `article_fetches`
- `extract-article-text` → `article_texts`
- `acquire-audio` → `audio_sources`
- `transcribe-audio` → `audio_transcriptions`
- `extract-spots` → `spot_mentions`, `presenters`, `article_presenters`
- `geocode-spots` → `spot_google_geocodes`, `article_spots`
- `export-data` → read-only (no DB table writes)

This ownership model is enforced in code/tests to keep retries deterministic and stage side-effects auditable.
