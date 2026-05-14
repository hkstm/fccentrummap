## 1. Schema and Repository Foundations

- [x] 1.1 Replace `Repository.InitSchema` DDL with relational v2 tables (`article_sources`, `article_fetches`, `article_texts`, `audio_sources`, `audio_transcriptions`, `spot_mentions`, `spot_google_geocodes`, `presenters`, `article_presenters`, `article_spots`) and required PK/FK/UNIQUE constraints.
- [x] 1.2 Remove legacy v1 table DDL and repository methods that are no longer part of the new write-ownership model.
- [x] 1.3 Add/adjust repository methods for latest-only `article_fetches` upsert by `article_source_id`.
- [x] 1.4 Add/adjust repository methods for single-row `article_texts` upsert by `article_fetch_id`.
- [x] 1.5 Add/adjust repository methods for `spot_mentions` writes from parsed extraction output (`place` + timestamp fields).
- [x] 1.6 Add/adjust repository methods for `presenters` upsert and `article_presenters` link writes.
- [x] 1.7 Add/adjust repository methods for `spot_google_geocodes` and `article_spots` writes owned by `geocode-spots`.

## 2. Stage Ownership Refactor

- [x] 2.1 Refactor `collect-article-urls` SQLite adapter/service to write only `article_sources`.
- [x] 2.2 Refactor `fetch-articles` SQLite adapter/service to write latest-only `article_fetches` rows (HTML only).
- [x] 2.3 Introduce/refactor `extract-article-text` stage to persist one `article_texts.cleaned_text` row per fetch.
- [x] 2.4 Refactor `acquire-audio` to derive video identity from `article_fetches.html` at read time and write only `audio_sources`.
- [x] 2.5 Refactor `transcribe-audio` to read from `audio_sources` and write only `audio_transcriptions`.
- [x] 2.6 Refactor `extract-spots` stage to write `spot_mentions`, `presenters`, and `article_presenters` as its owned tables.
- [x] 2.7 Refactor `geocode-spots` stage to write both `spot_google_geocodes` and `article_spots` in the same execution.
- [x] 2.8 Refactor `export-data` query path to read from `article_spots` + `spot_google_geocodes` + presenter links.

## 3. CLI and Contract Alignment

- [x] 3.1 Update unified scrape CLI subcommands to include `extract-article-text` and align command flow with the new stage sequence.
- [x] 3.2 Update stage mode validation matrix and command usage/help text for revised stage responsibilities.
- [x] 3.3 Ensure each stage delegates persistence through adapters consistent with single-writer table ownership.
- [x] 3.4 Remove obsolete command-layer assumptions tied to legacy tables and flows.

## 4. Tests and Verification

- [x] 4.1 Update repository schema tests to validate v2 tables, constraints, and foreign-key behavior.
- [x] 4.2 Add/adjust stage contract/parity tests to validate writer ownership mapping (each table written by one stage only).
- [x] 4.3 Add tests for latest-only fetch semantics (`article_fetches` upsert) and single-row text semantics (`article_texts` upsert).
- [x] 4.4 Add tests for `extract-spots` presenter materialization (`presenters` + `article_presenters`).
- [x] 4.5 Add tests for `geocode-spots` combined geocode + article-spot linking behavior.
- [x] 4.6 Run end-to-end pipeline on a fresh DB and verify all stage outputs and export payload correctness.

## 5. Cleanup and Documentation

- [x] 5.1 Remove dead code paths and models tied to deprecated v1 persistence tables.
- [x] 5.2 Update scraper/CLI docs to reflect destructive schema reset and new command sequence.
- [x] 5.3 Document canonical table-writer ownership mapping and stage responsibilities for maintainers.
- [x] 5.4 Add migration note that existing SQLite files are unsupported and must be reinitialized via `scrape init`.
