## 1. Article publication storage

- [x] 1.1 Add nullable article publication timestamp storage to the SQLite `article_sources` schema for fresh databases
- [x] 1.2 Add idempotent compatible-database schema evolution for existing databases missing the article publication timestamp column
- [x] 1.3 Implement publish-time parsing from fetched article HTML, preferring `article:published_time` and falling back to `datePublished`
- [x] 1.4 Populate article publication time when storing/upserting fetched article HTML
- [x] 1.5 Backfill missing article publication times from existing `article_fetches.html` during repository initialization or an equivalent idempotent repository path, failing with diagnostics when metadata is missing or unparseable

## 2. Export presenter ordering behavior

- [x] 2.1 Extend the export data query/scanning path to make stored article publication time available for internal ordering logic without changing exported JSON models
- [x] 2.2 Track the latest stored article publication time per presenter while building export data
- [x] 2.3 Update presenter sorting to order by latest associated article publication time descending
- [x] 2.4 Preserve deterministic tie-breaker ordering by `presenterName` ascending when latest publication times are equal
- [x] 2.5 Fail export with diagnostics when an exportable article has no stored publication time
- [x] 2.6 Ensure exported presenter and spot objects do not include publication-time fields

## 3. Tests and validation

- [x] 3.1 Add repository/schema tests covering fresh database creation and compatible existing-database column addition/backfill
- [x] 3.2 Add repository/export tests covering newest-first presenter ordering across multiple presenters
- [x] 3.3 Add tests for deterministic tie ordering, `article:published_time` preference/`datePublished` fallback parsing, and missing/invalid publish metadata failure behavior
- [x] 3.4 Run Go tests for the scraper package and re-export sample data if needed
