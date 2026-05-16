## Why

The frontend renders the presenter filter in the order provided by the static export. Alphabetical ordering is stable, but it does not prioritize currently relevant "Spots van" entries; ordering presenters by their latest article publication date makes recent Amsterdammers easier to find without changing the exported JSON shape.

## What Changes

- Store each article source's publication time in SQLite, populated from existing publish metadata in fetched article HTML.
- Sort the exported `presenters` array by each presenter's most recent associated article publication time, newest first.
- Preserve the existing presenter export schema; do not add publish-date fields to the JSON payload.
- Keep deterministic tie-breaker ordering by presenter name for equal publication times.
- Backfill publication times for existing compatible SQLite databases from stored fetched HTML, failing loudly if required publish metadata is missing or unparseable.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `sqlite-storage`: Article source storage is extended with a dedicated publication timestamp populated from fetched article HTML and backfilled for existing compatible databases.
- `static-site-data-dump`: Presenter collection ordering changes from generic deterministic ordering to latest-article-first ordering while preserving presenter values and JSON shape.
- `scraping-data-json-export`: Deterministic ordering requirements are refined to include publication-date-based presenter ordering with stable tie-breakers.

## Impact

- Affected SQLite schema/repository code: add nullable article publication timestamp storage and compatible existing-database backfill.
- Affected Go export path: repository export query/model-internal processing and export tests.
- Affected static output: order of entries in `presenters`; no new fields and no frontend schema change.
- No backend API or frontend component changes expected.
