## Context

The scraper pipeline currently has stages that write to multiple tables in a single run (for example, fetching article HTML and persisting extracted text in the same stage). This makes ownership boundaries unclear and increases operational complexity for retries, idempotency, and failure isolation.

This change is intentionally breaking: the SQLite schema will be rebuilt from scratch, with no migration compatibility requirements. The design must preserve normalized SQL modeling (no rich JSON persistence model) while enforcing a clear stage write-ownership contract: each table has exactly one writer stage, while a stage may write multiple tables it owns.

Constraints:
- Implementation remains in Go and SQLite (`modernc.org/sqlite`).
- Existing CLI entrypoint (`scrape`) remains the orchestration surface.
- External integrations (Colly crawl/fetch, Murmel transcription, Google geocoding) remain, but persistence boundaries change.

## Goals / Non-Goals

**Goals:**
- Enforce single-writer ownership per table across the scraper workflow (a table is written by exactly one stage).
- Introduce a new normalized relational schema with strict keys, foreign keys, and uniqueness for idempotency.
- Split current multi-write stage behavior into dedicated stage boundaries so write effects are isolated and auditable.
- Keep pipeline execution understandable from command ordering and table ownership alone.
- Preserve export capability for visualization data via relational joins.

**Non-Goals:**
- Backward compatibility with existing SQLite files.
- Building a JSON envelope/event-sourcing persistence model.
- Changing core external providers (Murmel/Google) or adding major dependencies.
- Redesigning frontend visualization payload shape in this phase.

## Decisions

### 1) Single-writer table ownership is a first-class contract
Each persistent table must have exactly one writer stage. A stage may write multiple tables when those tables are part of that stage’s owned responsibility. In this design, `extract-spots` writes `spot_mentions`, `presenters`, and `article_presenters`, and `geocode-spots` writes both `spot_google_geocodes` and `article_spots` for final spot linking.

Rationale:
- Improves failure isolation and retry behavior.
- Reduces unintended side effects and implicit coupling.
- Makes stage observability and testing simpler.

Alternatives considered:
- Keep current multi-table writes with better documentation: rejected because ambiguity remains operationally.
- Keep strict one-table-per-stage rule: rejected because it forces artificial stages without improving data ownership.
- Allow stage writes without table ownership constraints: rejected because it reintroduces ambiguous ownership and side effects.

### 2) Introduce relational v2 schema aligned to stage boundaries
The schema is reorganized so write targets map directly to stages, with canonical fields stored in exactly one table to avoid duplication drift. Article linkage is source-level (`article_source_id`) by design because `article_fetches` is latest-only (no fetch version history). Proposed core flow tables:
- `article_sources` (URL discovery ownership)
- `article_fetches` (fetched HTML ownership)
- `article_texts` (cleaned text extraction ownership)
- `audio_sources` (audio acquisition ownership; derives video identity from article HTML at read time)
- `audio_transcriptions` (transcription ownership)
- `spot_mentions` (LLM extraction ownership)
- `spot_google_geocodes` (Google geocoding ownership)
- `presenters` (materialized presenter dimension)
- `article_presenters` (materialized article↔presenter links)
- `article_spots` (materialized article↔geocoded-spot links)

Rationale:
- Keeps normalized SQL and clear data lineage.
- Enforces clear, auditable ownership: every table has one writer stage.
- Keeps `article_fetches` focused on raw fetched HTML; derived video identity is computed in `acquire-audio` instead of persisted in fetch storage.
- Allows pragmatic multi-table writes where necessary to keep presenter linkage and final spot linking consistent with extraction/geocoding outputs.
- Removes a separate materialization stage to reduce orchestration overhead.

Alternatives considered:
- Preserve existing table names and reshape logic only: rejected because existing semantics encode mixed responsibilities.

### 3) Split current commands into narrower write stages where needed
Stages that currently write multiple tables (especially around article fetch/text extraction and downstream entity/link writes) will be decomposed into explicit commands.

Expected command flow (high-level):
1. collect URLs
2. fetch article HTML
3. extract article text
4. acquire audio
5. transcribe audio
6. extract spot mentions (+ materialize presenters)
7. geocode mentions + link article spots
8. export data

Rationale:
- Command semantics and write semantics stay aligned.
- Easier to rerun only failed responsibilities.

Alternatives considered:
- Hide sub-steps behind one command while internally writing several tables: rejected because it obscures ownership and violates the operational intent.

### 4) Idempotency is enforced through natural uniqueness + upsert policy per table
Each owned table will define uniqueness constraints that represent natural identity (e.g., URL uniqueness, provider/language uniqueness per audio source, unique place per transcription), and each stage will use deterministic insert/upsert behavior.

Rationale:
- Safe retries without duplicate rows.
- Predictable stage re-execution behavior.

Alternatives considered:
- Pure append-only tables with periodic compaction: rejected as unnecessary complexity for this use case.

### 5) Export remains read-only over relational joins
`export-data` remains file-producing and does not own DB writes. It reads `article_spots` joined to `spot_google_geocodes` (and `spot_mentions`) plus presenter links.

Rationale:
- Keeps export deterministic and side-effect free.
- Maintains separation between data preparation and data rendering/export.

## Cleaned Article Text Model (rationale)

`article_texts` intentionally stores one normalized `cleaned_text` value per fetched article (`article_fetch_id`), instead of storing multiple text segments.

Rationale:
- The downstream consumer (`extract-spots`) ultimately needs a single cleaned article text block for prompt construction.
- A single-row model reduces schema/adapter complexity and avoids unnecessary row fan-out.
- It keeps extraction deterministic and idempotent with latest-only semantics (`article_fetch_id` is unique).

Expected `cleaned_text` content contract:
- Derived from `article_fetches.html` using deterministic parser-based extraction (no LLM).
- Main article body only (navigation, footer, unrelated boilerplate removed as much as possible by extractor).
- UTF-8 text, trimmed, with normalized whitespace/newlines suitable for direct prompt inclusion.
- Empty or low-quality extraction should fail stage validation rather than persisting unusable text.

Note on current implementation:
- Current code emits multiple extracted content items and later concatenates them before LLM use.
- v2 schema collapses this into one stored `cleaned_text` value to match actual downstream usage and simplify persistence.

## Proposed Schema (explicit DDL)

```sql
CREATE TABLE article_sources (
  article_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
  url TEXT NOT NULL UNIQUE,
  discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE article_fetches (
  article_fetch_id INTEGER PRIMARY KEY AUTOINCREMENT,
  article_source_id INTEGER NOT NULL UNIQUE
    REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
  html TEXT NOT NULL,
  fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- NOTE: latest-only fetch model (one row per article_source_id). Re-fetches upsert into the same row.

CREATE TABLE article_texts (
  article_text_id INTEGER PRIMARY KEY AUTOINCREMENT,
  article_fetch_id INTEGER NOT NULL UNIQUE
    REFERENCES article_fetches(article_fetch_id) ON DELETE CASCADE,
  cleaned_text TEXT NOT NULL,
  extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audio_sources (
  audio_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
  article_fetch_id INTEGER NOT NULL UNIQUE
    REFERENCES article_fetches(article_fetch_id) ON DELETE CASCADE,
  youtube_url TEXT NOT NULL,
  audio_format TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  audio_blob BLOB NOT NULL,
  byte_size INTEGER NOT NULL,
  acquired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audio_transcriptions (
  transcription_id INTEGER PRIMARY KEY AUTOINCREMENT,
  audio_source_id INTEGER NOT NULL
    REFERENCES audio_sources(audio_source_id) ON DELETE CASCADE,
  provider TEXT NOT NULL,
  language TEXT NOT NULL,
  http_status INTEGER NOT NULL,
  response_json TEXT NOT NULL CHECK(json_valid(response_json)),
  response_byte_size INTEGER NOT NULL,
  error_message TEXT,
  transcribed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(audio_source_id, provider, language)
);

CREATE TABLE spot_mentions (
  spot_mention_id INTEGER PRIMARY KEY AUTOINCREMENT,
  transcription_id INTEGER NOT NULL
    REFERENCES audio_transcriptions(transcription_id) ON DELETE CASCADE,
  place TEXT NOT NULL,
  sentence_start_timestamp REAL,
  original_sentence_start_timestamp REAL,
  refined_sentence_start_timestamp REAL,
  extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(transcription_id, place)
);

CREATE TABLE spot_google_geocodes (
  spot_google_geocode_id INTEGER PRIMARY KEY AUTOINCREMENT,
  spot_mention_id INTEGER NOT NULL UNIQUE
    REFERENCES spot_mentions(spot_mention_id) ON DELETE CASCADE,
  google_place_id TEXT,
  latitude REAL NOT NULL,
  longitude REAL NOT NULL,
  formatted_address TEXT,
  status TEXT NOT NULL,
  geocoded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- NOTE: latest-successful-only model. One geocode row per mention; no retry history table.

CREATE TABLE presenters (
  presenter_id INTEGER PRIMARY KEY AUTOINCREMENT,
  presenter_name TEXT NOT NULL UNIQUE,
  materialized_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE article_presenters (
  article_source_id INTEGER NOT NULL
    REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
  presenter_id INTEGER NOT NULL
    REFERENCES presenters(presenter_id) ON DELETE CASCADE,
  linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (article_source_id, presenter_id)
);

CREATE TABLE article_spots (
  article_source_id INTEGER NOT NULL
    REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
  spot_google_geocode_id INTEGER NOT NULL
    REFERENCES spot_google_geocodes(spot_google_geocode_id) ON DELETE CASCADE,
  linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (article_source_id, spot_google_geocode_id)
);
```

Table writer ownership:
- `article_sources` ← `collect-article-urls`
- `article_fetches` ← `fetch-articles`
- `article_texts` ← `extract-article-text`
- `audio_sources` ← `acquire-audio`
- `audio_transcriptions` ← `transcribe-audio`
- `spot_mentions` ← `extract-spots`
- `presenters` ← `extract-spots`
- `article_presenters` ← `extract-spots`
- `spot_google_geocodes` ← `geocode-spots`
- `article_spots` ← `geocode-spots`
- `export-data` writes no DB tables

## Risks / Trade-offs

- **[Risk] Multi-table stages can become overloaded and harder to maintain** → **Mitigation:** keep ownership boundaries explicit and ensure each written table has a single writer stage.
- **[Risk] Stage explosion increases command count and operator burden** → **Mitigation:** consolidate presenter writes into `extract-spots` and consolidate spot-link materialization into `geocode-spots`.
- **[Risk] Incorrect FK/unique design can block legitimate rows or allow duplicates** → **Mitigation:** specify constraints per table in specs and validate with repository/integration tests.
- **[Risk] Transition churn across adapters/services is broad** → **Mitigation:** implement table/stage changes incrementally with per-stage contract tests.
- **[Risk] Existing automation/scripts assume old stage effects** → **Mitigation:** update CLI docs and fail fast with clear errors when old assumptions are used.
- **[Trade-off] More intermediate tables improve ownership clarity but add join complexity** → **Mitigation:** keep table purposes narrow and document query patterns for export/debugging.

## Migration Plan

1. Introduce v2 schema DDL and remove v1 schema definitions in `InitSchema`.
2. Implement repository methods grouped by new stage ownership tables.
3. Refactor pipeline adapters/services stage-by-stage to enforce single-writer ownership per table.
4. Add/adjust commands for split responsibilities (extract text + presenter materialization in `extract-spots` + spot link materialization in `geocode-spots`).
5. Update tests (repository, stage parity/contract tests) to new ownership model.
6. Run end-to-end pipeline on a fresh DB and verify each table is written by only its designated writer stage.
7. Update docs/usage; old DB files are unsupported and should be recreated via `scrape init`.

Rollback strategy:
- Code rollback to pre-change commit is possible, but DB rollback is not required because this is a destructive schema reset. Reinitialize DB on rollback.

## Open Questions

- None currently.
