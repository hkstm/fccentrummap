## 1. Schema and Source Model

- [x] 1.1 Add `SpotSource` domain/request value with normalization for empty source to `gemini-direct` and validation for `gemini-direct|transcript`
- [x] 1.2 Add idempotent SQLite schema initialization for `gemini_direct_spot_mentions`, `gemini_direct_spot_google_geocodes`, `gemini_direct_article_spots`, and `gemini_direct_spot_corrections`
- [x] 1.3 Add repository tests proving fresh initialization creates Gemini-direct tables, constraints, indexes/uniqueness, and foreign keys
- [x] 1.4 Add repository tests proving existing compatible databases initialize Gemini-direct tables without dropping existing transcript data

## 2. Gemini-Direct Extraction Persistence

- [x] 2.1 Update Gemini-direct structured response types and prompt contract to include article/video validation context, human-readable evidence, optional confidence, and model metadata
- [x] 2.2 Implement validation for article URL, YouTube URL, non-empty place names, non-negative timestamps, in-range confidence values, and trimmed presenter names before persistence
- [x] 2.3 Add repository methods to upsert Gemini-direct spot mentions by `(article_source_id, place)` with YouTube URL, timestamp, evidence, confidence, model, and extraction time
- [x] 2.4 Add repository or adapter methods to upsert presenter records and article-presenter links from valid Gemini-direct responses without requiring audio/transcription rows
- [x] 2.5 Update `extract-spots-gemini-direct` SQLite adapter/service flow to persist valid parsed responses directly after parsing while still writing deterministic diagnostic artifacts
- [x] 2.6 Add tests for accepted Gemini responses, invalid context rejection, invalid candidate rejection, presenter linkage, confidence persistence, and idempotent reruns

## 3. Source-Neutral Repository and Adapter Ports

- [x] 3.1 Define source-neutral DTOs for spot mentions to geocode, source-selected geocode writes, export rows, correction target lookup, and correction upserts
- [x] 3.2 Add repository methods to list ungeocoded spot mentions for transcript and Gemini-direct sources through the same source-neutral shape
- [x] 3.3 Add repository methods to upsert geocode rows and article links for the selected source, routing transcript writes to existing tables and Gemini writes to Gemini-specific tables
- [x] 3.4 Add repository methods to export rows for the selected source, routing transcript reads to existing joins and Gemini reads to Gemini-specific joins
- [x] 3.5 Add repository methods to resolve correction targets and upsert corrections for the selected source using source-specific correction tables

## 4. Geocode Stage Source Selection

- [x] 4.1 Add `SpotSource` to `geocode-spots` request DTOs, validation, CLI flag parsing, and command diagnostics with Gemini-direct as the default
- [x] 4.2 Update geocode stage ports and SQLite adapter to consume selected source when listing mentions and writing geocode/article-link rows
- [x] 4.3 Replace the SQLite-mode rejection path for `geocode-spots` with source-selected SQLite geocoding while preserving explicit file-mode artifact behavior
- [x] 4.4 Ensure `geocode-spots --export-json` passes the same selected source into inline export so geocoding and export cannot use different sources in one run
- [x] 4.5 Add geocode tests for Gemini default behavior, explicit transcript fallback, unsupported source validation, source-specific writes, and file-mode preservation

## 5. Export and Correction Source Selection

- [x] 5.1 Add `SpotSource` to `export-data` request DTOs, validation, CLI flag parsing, and command diagnostics with Gemini-direct as the default
- [x] 5.2 Update export SQLite adapter to read selected-source export rows and corrections while keeping the JSON output schema unchanged
- [x] 5.3 Add export tests for Gemini default output, transcript fallback output, unsupported source validation, unchanged JSON shape, and source-specific correction application
- [x] 5.4 Add `SpotSource` to `correct-spot` command flow so default corrections resolve/write Gemini-specific rows and `--spot-source transcript` preserves existing transcript behavior
- [x] 5.5 Add correction tests for Gemini target lookup, Gemini correction upsert, transcript fallback, same-string spot ID isolation across sources, and unsupported source validation

## 6. Comparison, Ownership, and Regression Coverage

- [x] 6.1 Update `compare-spot-extractions` only as needed to read Gemini-direct rows from SQLite while preserving its explicit transcript-vs-Gemini comparison semantics
- [x] 6.2 Update stage writer ownership documentation/tests so Gemini-direct extraction owns `gemini_direct_spot_mentions`, geocode owns Gemini geocode/link tables, and correction flow owns Gemini correction rows
- [x] 6.3 Add or update end-to-end SQLite pipeline tests covering Gemini extraction persistence, source-selected geocoding, source-selected export, and explicit transcript rollback/fallback
- [x] 6.4 Update user-facing CLI help or README documentation for `--spot-source gemini-direct|transcript`, Gemini-direct default behavior, and transcript fallback usage
- [x] 6.5 Run `go test ./...` from `scraper/` and `./.pi/bin/openspec validate integrate-gemini-extracted-spots --strict`
