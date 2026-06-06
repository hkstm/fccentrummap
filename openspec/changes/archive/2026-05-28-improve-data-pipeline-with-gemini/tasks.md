## 1. Schema and Data Reset

- [x] 1.1 Update SQLite initialization/migration code to remove legacy transcription-path tables from the active schema and keep only article, presenter, Gemini-direct, geocode, article-link, and correction tables.
- [x] 1.2 Add nullable `address` storage to `gemini_direct_spot_mentions` and nullable `primary_type_display_name` storage to `gemini_direct_spot_google_geocodes`.
- [x] 1.3 Implement or document the operator reset path that deletes existing Gemini-derived spot mentions, geocodes, article links, and corrections while preserving article sources/fetches.
- [x] 1.4 Update schema tests or smoke checks to verify fresh initialization creates the Gemini-first schema and does not require legacy transcript tables.

## 2. Gemini Direct Extraction

- [x] 2.1 Update Gemini response DTOs, parsing, validation, and persisted artifacts to carry nullable spot `address` values.
- [x] 2.2 Update the Gemini prompt to request optional addresses only when reasonably confident and to explain transition-card subtitle handling.
- [x] 2.3 Update the Gemini prompt to prefer canonical presenter spelling from article/video titles when appropriate.
- [x] 2.4 Persist trimmed non-empty address values to `gemini_direct_spot_mentions` and reject whitespace-only address values during validation.
- [x] 2.5 Update Gemini extraction tests/fixtures to cover address present, address omitted/null, canonical presenter spelling, and malformed address validation.

## 3. Geocoding

- [x] 3.1 Update Gemini spot geocoding row queries to read stored `address` values with each ungeocoded mention.
- [x] 3.2 Build Google Places Text Search `textQuery` as `<place>, <address>` when address is available and `<place>, Amsterdam` otherwise.
- [x] 3.3 Request and parse `places.primaryTypeDisplayName` from Google Places Text Search and persist it to Gemini geocode rows.
- [x] 3.4 Remove transcript source selection paths from the active geocode command behavior so the command operates on Gemini-direct data.
- [x] 3.5 Add geocoding unit tests for address-aware query construction, fallback query construction, field mask coverage, and primary type display name persistence.

## 4. JSON Export and Frontend Data Contract

- [x] 4.1 Update `export-data` to read Gemini-direct geocodes, links, corrections, and presenter attribution without transcript source selection.
- [x] 4.2 Include each spot's primary type display name in exported spot records when stored.
- [x] 4.3 Remove the exported top-level `presenters` array while keeping presenter names available on spot records.
- [x] 4.4 Update frontend TypeScript data types and data loading to accept exports without a top-level `presenters` array.
- [x] 4.5 Derive unique presenter filter options from the exported `spots` array and de-duplicate presenter names deterministically.
- [x] 4.6 Update export and frontend tests to cover empty exports, primary type display names, missing top-level presenters, and derived presenter filters.

## 5. Legacy Transcription Cleanup

- [x] 5.1 Remove legacy transcription extraction code, DTOs, services, repositories, and tests that are no longer used by the Gemini-first pipeline.
- [x] 5.2 Remove CLI subcommands and documentation references for legacy transcript stages from the current pipeline surface.
- [x] 5.3 Remove old transcript-derived table access from current downstream code paths, including `spot_mentions`, `spot_google_geocodes`, `article_spots`, and `spot_corrections`.
- [x] 5.4 Update project docs and examples to show the current rerun sequence: init/reset, collect/fetch as needed, `extract-spots-gemini-direct`, `geocode-spots`, and `export-data`.

## 6. Verification and Reprocessing

- [x] 6.1 Run Go tests for the scraper module and fix regressions.
- [x] 6.2 Run frontend tests/type checks for the viz module and fix regressions.
- [x] 6.3 Run a local pipeline smoke test against a disposable database to verify extraction artifacts, address-aware geocoding inputs, and export shape.
- [x] 6.4 Drop old generated Gemini/geocoding data from the target database using the reset path.
- [x] 6.5 Re-run extraction and geocoding from scratch and regenerate the frontend JSON export.
- [x] 6.6 Review regenerated geocoding/export results for obvious address or coordinate regressions before considering the change complete.
