## Context

The existing transcript pipeline writes extracted places to `spot_mentions`, then `geocode-spots` reads ungeocoded mentions, writes `spot_google_geocodes`/`article_spots`, and `export-data` reads those rows to build the frontend JSON. The existing table you were thinking of is `spot_mentions`: it stores one extracted place mention with timestamp fields and is the source row that geocoding consumes.

The Gemini-direct pipeline is intentionally based on article URL + YouTube URL context, not downloaded audio or transcript rows. It should not require `audio_sources` or `audio_transcriptions` to exist. The desired model is not a separate promotion workflow: Gemini-direct extraction should write successful parsed output directly to SQLite into a Gemini-specific table, and stages that currently consume transcript-derived spot data should accept a flag to consume Gemini-derived spot data instead.

## Goals / Non-Goals

**Goals:**

- Persist successful Gemini-direct spot candidates directly to SQLite as part of `extract-spots-gemini-direct`.
- Add a Gemini-derived spot mention table that stores the fields downstream stages actually need while linking directly to `article_sources` and the YouTube URL instead of requiring audio or transcription rows.
- Add explicit source-selection flags, such as `--spot-source gemini-direct|transcript`, to downstream stages that consume spot/geocode data.
- Make Gemini-derived data the default downstream source; operators pass `--spot-source transcript` to fall back to the older transcript-derived path.
- Keep the exported JSON schema unchanged regardless of selected source.
- Preserve deterministic artifact writing as diagnostics/review output, not as the integration boundary.
- Persist every valid Gemini candidate regardless of confidence score, while storing the confidence value for later analysis/filtering.

**Non-Goals:**

- Adding a separate operator promotion stage after Gemini extraction.
- Requiring fake audio or transcript rows solely to fit Gemini data into transcript-shaped tables.
- Building a manual candidate review UI.
- Automatically merging or deduplicating transcript and Gemini outputs into a single canonical source.

## Decisions

### Decision 1: Gemini-direct extraction writes DB records directly

`extract-spots-gemini-direct` will persist each valid parsed response into SQLite after the model response is parsed and validated. Artifacts may still be written for observability, replay, and debugging, but downstream integration will not depend on reading artifact files.

Rationale: this matches the intended pipeline shape: Gemini extraction is an alternate extraction backend whose output can feed the same downstream stages as transcript extraction.

Alternatives considered:

- Add a separate `integrate-gemini-extracted-spots` command that reads artifacts and promotes them later. This preserves a review checkpoint, but it adds an extra operational step the desired flow does not need.
- Make downstream stages read Gemini artifact files directly. This avoids schema work, but couples stage execution to artifact paths and bypasses database consistency guarantees.

### Decision 2: Add source-specific Gemini tables that mirror the transcript source shape

Add `gemini_direct_spot_mentions` as the Gemini equivalent of `spot_mentions`. Because existing geocode tables have foreign keys to `spot_mentions`, also add Gemini-specific geocode/link tables that mirror `spot_google_geocodes` and `article_spots`. To avoid exposing source prefixes in exported spot IDs, also add Gemini-specific correction storage that mirrors `spot_corrections`.

The repository should expose source-neutral DTOs to stage services, while SQLite adapters route to the transcript or Gemini table set based on `SpotSource`.

#### Proposed SQLite schema

```sql
CREATE TABLE IF NOT EXISTS gemini_direct_spot_mentions (
    gemini_direct_spot_mention_id INTEGER PRIMARY KEY AUTOINCREMENT,
    article_source_id INTEGER NOT NULL
        REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
    youtube_url TEXT NOT NULL,
    place TEXT NOT NULL,
    youtube_timestamp_seconds REAL CHECK (youtube_timestamp_seconds IS NULL OR youtube_timestamp_seconds >= 0),
    evidence TEXT,
    confidence REAL CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1)),
    model TEXT,
    extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(article_source_id, place)
);

CREATE TABLE IF NOT EXISTS gemini_direct_spot_google_geocodes (
    gemini_direct_spot_google_geocode_id INTEGER PRIMARY KEY AUTOINCREMENT,
    gemini_direct_spot_mention_id INTEGER NOT NULL UNIQUE
        REFERENCES gemini_direct_spot_mentions(gemini_direct_spot_mention_id) ON DELETE CASCADE,
    google_place_id TEXT,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    formatted_address TEXT,
    primary_type TEXT,
    primary_type_display_name TEXT,
    status TEXT NOT NULL,
    geocoded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS gemini_direct_article_spots (
    article_source_id INTEGER NOT NULL
        REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
    gemini_direct_spot_google_geocode_id INTEGER NOT NULL
        REFERENCES gemini_direct_spot_google_geocodes(gemini_direct_spot_google_geocode_id) ON DELETE CASCADE,
    linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (article_source_id, gemini_direct_spot_google_geocode_id)
);

CREATE TABLE IF NOT EXISTS gemini_direct_spot_corrections (
    spot_id TEXT PRIMARY KEY CHECK (trim(spot_id) <> ''),
    spot_name TEXT,
    place_id TEXT,
    latitude REAL,
    longitude REAL,
    youtube_timestamp_seconds INTEGER CHECK (youtube_timestamp_seconds IS NULL OR youtube_timestamp_seconds >= 0),
    hidden INTEGER NOT NULL DEFAULT 0 CHECK (hidden IN (0, 1)),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        ((place_id IS NULL OR trim(place_id) = '') AND latitude IS NULL AND longitude IS NULL)
        OR (place_id IS NOT NULL AND trim(place_id) <> '' AND latitude IS NOT NULL AND longitude IS NOT NULL)
    )
);
```

Notes:

- `gemini_direct_spot_mentions` stores the fields downstream stages need: place, one YouTube timestamp, extraction time, and an ID usable by geocoding/export.
- `article_source_id` is stored directly because Gemini-direct extraction is article/video-context based rather than transcription based.
- `youtube_url` is stored directly because Gemini-direct does not require downloaded audio, `audio_sources`, or `audio_transcriptions`; the URL is part of the Gemini request identity and lets export build timestamped YouTube links without joining through transcript/audio tables.
- Gemini-direct extraction is not a two-pass transcript refinement flow, so it stores a single `youtube_timestamp_seconds` value instead of transcript-specific `sentence_start_timestamp`, `original_sentence_start_timestamp`, and `refined_sentence_start_timestamp` fields.
- `evidence`, `confidence`, and `model` preserve minimal Gemini provenance without changing downstream export shape. Artifact paths remain filesystem diagnostics and are not stored on per-spot DB rows. The Gemini response contract should request one human-readable evidence/rationale field rather than separate `evidence` and `notes` fields.
- `confidence` is persisted for all candidates when Gemini provides it, but the first version does not filter or reject candidates based on confidence threshold.
- `UNIQUE(article_source_id, place)` mirrors the existing `UNIQUE(transcription_id, place)` behavior and makes reruns idempotent.
- Gemini exported spot IDs can keep the same public shape as transcript IDs, `<article_source_id>:<gemini_direct_spot_mention_id>`, because Gemini corrections are stored in `gemini_direct_spot_corrections` instead of sharing `spot_corrections`.

Rationale: a parallel source table keeps transcript and Gemini extraction outputs distinct while giving downstream stages a similar row shape to consume.

Alternatives considered:

- Store Gemini output in `spot_mentions` with a synthetic transcription row. This reuses existing geocode/export joins, but makes `audio_transcriptions` contain non-transcription data and obscures provenance.
- Add an `extraction_source` column to `spot_mentions`. This avoids parallel tables, but requires broader migration of uniqueness and foreign-key assumptions and mixes sources in a table that currently implies transcript provenance.
- Generalize `spot_google_geocodes` with nullable transcript/Gemini mention IDs and source checks. This is cleaner long-term, but more invasive than adding parallel Gemini geocode/link tables.

### Decision 3: Add source selection only to stages that consume spot/geocode sources

Use `--spot-source gemini-direct|transcript` as the operator-facing selector. Default is `gemini-direct`; `--spot-source transcript` is the explicit fallback to the older transcript-based tables.

Downstream command/stage impact:

| Command/stage | Needs `--spot-source`? | Why |
| --- | --- | --- |
| `geocode-spots` in SQLite mode | Yes | Default behavior reads `gemini_direct_spot_mentions` and writes `gemini_direct_spot_google_geocodes`/`gemini_direct_article_spots`; `--spot-source transcript` falls back to `spot_mentions` and `spot_google_geocodes`/`article_spots`. |
| `geocode-spots --export-json` | Yes, same flag | The inline export must use the same selected source as the geocode run to avoid geocoding one source and exporting another. |
| `export-data` | Yes | Default behavior joins the Gemini table set and emits the existing JSON shape; `--spot-source transcript` falls back to `article_spots` → `spot_google_geocodes` → `spot_mentions`. |
| `correct-spot` | Yes | It edits by exported `spotId`. Default behavior should resolve the ID against Gemini tables and write `gemini_direct_spot_corrections`; `--spot-source transcript` falls back to the existing transcript lookup and `spot_corrections`. This keeps public spot IDs in the existing `<article_source_id>:<mention_id>` shape without exposing a `gemini-direct`/`g:` prefix. |
| `compare-spot-extractions` | No generic source flag | It is a comparison tool with explicit semantics: transcript baseline versus Gemini-direct output. It may be updated to load Gemini-direct rows from SQLite, but it should not become a generic source-selected pipeline stage. |
| `extract-spots`, `extract-spots-gemini-direct` | No downstream source flag | These are producers. `extract-spots` writes transcript source rows; `extract-spots-gemini-direct` writes Gemini source rows. |
| Collection/fetch/text/audio/transcription stages | No | They do not consume spot/geocode source tables. |

Rationale: only stages that read or write source-specific spot/geocode tables need source selection. This keeps the rest of the CLI stable and avoids unnecessary flags on unrelated stages.

Alternatives considered:

- Add separate commands for every Gemini downstream stage. This is explicit, but duplicates command/service logic and grows the CLI surface unnecessarily.
- Merge transcript and Gemini candidates before geocoding. This makes export simpler, but removes the operator's ability to compare or select extraction sources independently.

### Decision 4: Use source-neutral repository/service DTOs with source-specific SQL

`geocode-spots` should not duplicate geocoding orchestration. Instead, the SQLite adapter/repository should provide source-neutral methods such as:

- `ListSpotMentionsWithoutGeocode(source SpotSource) ([]SpotMentionForGeocode, error)`
- `UpsertSpotGoogleGeocodeAndLinkArticleSpot(source SpotSource, mentionID int64, ..., articleSourceID int64, ...) (int64, error)`
- `ExportDataForSource(source SpotSource) (*models.ExportData, error)`
- `GetSpotCorrectionTarget(source SpotSource, spotID string)` that queries the selected source's spot/geocode/correction tables
- `UpsertSpotCorrection(source SpotSource, correction models.SpotCorrection)` that writes to `gemini_direct_spot_corrections` by default or `spot_corrections` for transcript fallback

The service-level DTO should include a `SpotSource` field for source-consuming stages. Validation should accept only `gemini-direct` and `transcript`; empty source normalizes to `gemini-direct`.

Rationale: source-neutral DTOs keep business services simple while preserving database integrity with source-specific foreign keys.

### Decision 5: Validate article/video identity before persistence

Before writing Gemini-derived rows, the stage will verify that the parsed response belongs to the expected article source and video context:

- the article source exists;
- the stored article URL matches the parsed response context;
- the stored YouTube URL matches the parsed response context when available;
- spots have non-empty place names and non-negative `youtube_timestamp_seconds` values when timestamps are present;
- confidence values are persisted when present and must be within the valid Gemini response range, but no minimum confidence threshold is applied;
- presenter names are trimmed and empty presenters are ignored;
- malformed responses or responses with diagnostics are not persisted as accepted spots.

Rationale: direct DB writes make validation more important. Stale or mismatched responses must not pollute the selected Gemini source.

## Risks / Trade-offs

- **Parallel tables add query paths** → Mitigate by keeping service DTOs source-neutral and isolating table selection in adapters/repository methods.
- **Parallel source tables add intentional schema duplication** → Accept this for now because it keeps Gemini and transcript flows independent and avoids risky migrations. Mitigate scope by duplicating the source-specific spot/geocode/link/correction tables while reusing shared article, audio, presenter, and export structures.
- **Operators may export the wrong source** → Mitigate with explicit `--spot-source` CLI output, a documented Gemini default, and an explicit `--spot-source transcript` fallback for the old path.
- **Transcript and Gemini outputs may diverge or duplicate spots** → Mitigate by keeping sources selectable rather than auto-merging in this change.
- **Existing foreign keys constrain reuse of geocode tables** → Mitigate with source-specific geocode/link tables for this iteration.
- **Correction ID collisions** → Mitigate by using source-specific correction tables. Gemini and transcript exports may both use `<article_source_id>:<mention_id>` IDs because correction lookup/writes are scoped by `--spot-source`, defaulting to Gemini.

## Migration Plan

1. Add idempotent SQLite schema initialization for `gemini_direct_spot_mentions`, `gemini_direct_spot_google_geocodes`, `gemini_direct_article_spots`, and `gemini_direct_spot_corrections`.
2. Add repository methods to validate article/video identity and upsert Gemini-direct spot mentions/presenter links from parsed Gemini responses.
3. Change `extract-spots-gemini-direct` SQLite mode to persist valid parsed responses directly after response parsing.
4. Add `SpotSource` request fields, CLI flags, normalization, and adapter routing for `geocode-spots` SQLite mode; pass the same source through `geocode-spots --export-json`.
5. Add `SpotSource` request fields, CLI flags, and repository export queries for `export-data` while keeping output JSON unchanged.
6. Update correction target lookup and correction writes to be source-selected: Gemini defaults use `gemini_direct_spot_corrections.spot_id`, transcript fallback uses `spot_corrections.spot_id`, and both sources can keep the public `<article_source_id>:<mention_id>` spot ID shape.
7. Add tests for schema initialization, Gemini extraction persistence, idempotent reruns, source-selected geocoding, source-selected export, source-aware correction lookup, Gemini default behavior, and explicit transcript fallback behavior.

Rollback strategy: because the transcript path remains unchanged, rollback can pass `--spot-source transcript` on source-consuming stages, ignore or drop the Gemini-direct tables, and continue using the old transcript source.

## Open Questions

None at this time.
