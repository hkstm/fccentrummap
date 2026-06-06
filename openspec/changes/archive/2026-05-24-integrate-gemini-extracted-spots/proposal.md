## Why

Gemini-direct spot extraction can produce structured spot candidates, but those candidates do not yet enter the database shape consumed by the rest of the scraper pipeline. Writing Gemini-derived spots into a dedicated database table and letting downstream stages select that source enables the Gemini path to run end-to-end without manual artifact conversion.

## What Changes

- Change the Gemini-direct extraction path so successful parsed responses are persisted directly to SQLite, not only to review artifacts.
- Add a dedicated Gemini-derived spot mention table that mirrors the downstream shape of the existing transcript-based `spot_mentions` data while linking directly to the article/video context used by Gemini.
- Add stage flags for downstream consumers so operators can choose Gemini-derived spots or fall back to transcript-derived spots when geocoding and exporting.
- Make Gemini-derived spots the default source for downstream stages; require an explicit transcript source flag to use the older transcript-based path.
- Validate Gemini-derived candidates before persistence, including article/video identity, required place names, timestamps, and presenter data when available.

## Capabilities

### New Capabilities
- `gemini-extracted-spot-integration`: Defines how Gemini-direct responses are written into source-selectable database tables and consumed by downstream pipeline stages.

### Modified Capabilities
- `gemini-direct-spot-extraction`: Gemini-direct extraction now writes valid parsed spot output directly to SQLite while retaining artifacts only as diagnostics/review aids.
- `sqlite-storage`: Adds durable SQLite storage for Gemini-derived spot mentions and any source-specific downstream link/geocode records required by the existing relational model.
- `pipeline-layered-architecture`: Adds source selection to stage ports/adapters so downstream stages can consume transcript or Gemini spot sources without duplicating orchestration logic.
- `google-places-text-search-geocoding`: Allows the geocode stage to geocode either transcript-derived or Gemini-derived spot mentions based on an explicit operator flag.
- `scraping-data-json-export`: Allows the export stage to export either transcript-derived or Gemini-derived geocoded spots based on an explicit operator flag while keeping the JSON shape unchanged.

## Impact

- Affects the Go Gemini-direct extraction stage, repository schema, stage adapters, and CLI normalization/validation.
- Adds SQLite tables or table abstractions for Gemini-derived spot mentions and downstream geocode/article links where existing foreign keys require source-specific storage.
- Adds source-selection flags to downstream stages that consume spot mentions or geocoded spot links.
- Downstream frontend JSON shape remains unchanged; only the selected database source changes.
