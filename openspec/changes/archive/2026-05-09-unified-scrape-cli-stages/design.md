## Context

The scraper pipeline currently exists as multiple focused CLIs (`scraper`, `transcribe-audio`, `extract-spots-dry-run`, `geocode-place`, `export`, etc.) with uneven orchestration and inconsistent persistence surfaces. This change introduces a single entrypoint (`scrape`) with stage subcommands and a shared execution contract. Constraints from proposal: backwards compatibility is not required, database reset is acceptable, but feature parity with current behavior is required.

## Goals / Non-Goals

**Goals:**
- Provide one CLI entrypoint with explicit stage subcommands.
- Enforce stage boundaries: each stage performs at most one external expensive I/O category.
- Make SQLite the default input/output path for every stage.
- Provide file mode for every stage: read stage input artifact file(s), write stage output artifact file(s).
- Preserve all current capabilities (crawl/fetch, text extraction, YouTube audio, Murmel transcription, Gemma extraction, geocoding, export) under the new command surface.
- Make stage outputs composable and inspectable for debugging/replay.
- Remove legacy command entrypoints and legacy-only code paths once parity is verified.

**Non-Goals:**
- Preserving old command names or exact legacy flags.
- Finalizing every internal stage implementation in this change (scaffolding-first is acceptable where code is pending).
- Introducing new external providers beyond existing ones.

## Decisions

1. **Single entrypoint with stage subcommands**
   - Decision: Create `scraper/cmd/scrape` as the orchestration surface with subcommands:
     - `init`
     - `collect-article-urls`
     - `fetch-articles`
     - `acquire-audio`
     - `transcribe-audio`
     - `extract-spots`
     - `geocode-spots`
     - `export-data`
   - Rationale: separates concerns by expensive I/O type and makes partial pipeline runs explicit.
   - Alternatives:
     - Keep one monolithic `run` command: rejected (harder to debug/replay).
     - Keep many binaries: rejected (fragmented UX).

2. **External I/O boundary by stage**
   - Decision: each stage may call only one external expensive service class:
     - fccentrum.nl (collect/fetch)
     - YouTube/yt-dlp (audio)
     - Murmel (transcription)
     - Gemma (spot extraction)
     - Google Places (geocoding)
   - Rationale: predictable retries, cost control, and easier failure isolation.
   - Alternative: mixed multi-service stages; rejected due to harder observability/recovery.

3. **Unified I/O contract: SQLite default + file mode optional**
   - Decision: all subcommands accept an I/O mode with SQLite default and file override.
   - Proposed interface:
     - `--io sqlite|file` (default `sqlite`)
     - SQLite mode uses `--db-path`
     - File mode requires explicit stage-specific `--in` path(s)
     - File mode does **not** require `--out`; output path is derived deterministically from input identity + stage naming rules
     - No automatic latest-artifact discovery
   - Naming contract:
     - Every artifact carries a stable identity key (e.g., article slug/url key, audio-source key, transcription key, spot key)
     - Downstream stage outputs MUST preserve that identity key and append stage/type suffixes
     - If an external API returns a canonical identifier, include it as metadata and optional filename suffix, but keep the original pipeline identity stable
   - Rationale: consistent UX and stage replay in CI/manual workflows, with deterministic and auditable file inputs/outputs and chainable filenames across stages.
   - Alternative: ad hoc per-command flags; rejected for inconsistency.

4. **Stage artifact directory layout under `data/`**
   - Decision: introduce stage-oriented artifact folders:
     - `data/stages/collect-article-urls/`
     - `data/stages/fetch-articles/`
     - `data/stages/acquire-audio/`
     - `data/stages/transcribe-audio/`
     - `data/stages/extract-spots/`
     - `data/stages/geocode-spots/`
     - `data/stages/export-data/`
   - Filename convention is identity-first (derived from input artifact identity), with stage/type suffix and optional timestamp for retries/versioning.
   - Example pattern: `<identity>__<stage>__<table-or-payload-type>.json` (or audio extension for blobs).
   - Rationale: understandable by stage while retaining DB traceability and deterministic cross-stage chaining.

5. **Schema evolution allowed with reset-first migration**
   - Decision: allow DB schema changes needed for clean stage contracts; prefer additive schema with optional reset tooling, and document non-compatibility.
   - Rationale: user explicitly allows reset and values maintainable model over compatibility.
   - Alternative: strict compatibility migrations; rejected for added complexity with low value here.

6. **Init-time environment validation**
   - Decision: `scrape init` SHALL validate required API environment variables up front and fail fast with a clear missing-variable report.
   - Required API env vars (from current code paths):
     - Murmel: `MURMEL_API_KEY`
     - Google Places geocoding: `GOOGLE_MAPS_API_KEY`
     - Gemma/Gemini extraction API key: at least one of `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GOOGLE_GENERATIVE_LANGUAGE_API_KEY`
   - Non-required/optional env vars:
     - `GEMMA_MODEL` (has CLI/default fallback)
     - `GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT`, `GOOGLE_PLACES_TEXT_SEARCH_ENDPOINT` (endpoint overrides)
   - Rationale: prevent expensive partial runs from failing late after earlier stages have already consumed time/cost.

7. **Geocode persistence scope (explicitly deferred)**
   - Decision: in this change, `geocode-spots` does **not** write final `spots` rows in SQLite mode; geocode output remains file-artifact based until dedicated SQLite write integration is implemented in a follow-up change.
   - Rationale: keeps this change scoped to unified CLI scaffolding/refactor without introducing new SQLite persistence operations.

8. **Unsupported mode behavior MUST fail explicitly**
   - Decision: if a stage is invoked in a mode that is not supported yet, the CLI SHALL fail with a clear non-zero error and actionable message (no silent no-op).
   - Example: `scrape geocode-spots` with default `--io sqlite` must fail with a message like: "geocode-spots does not support --io sqlite yet; use --io file --in <path>".
   - Rationale: prevents ambiguity and accidental false-success runs.

9. **Legacy removal policy**
   - Decision: temporary wrappers/adapters may be used only during implementation, but final merge must remove legacy binaries/entrypoints and legacy-only code paths.
   - Rationale: avoid long-term dual maintenance and enforce the unified CLI as the single supported interface.

## Stage I/O Contract (SQLite-first, file fallback where not implemented)

Intent for this change:
- Prefer SQLite input/output for every stage where behavior already exists.
- If a stage's SQLite write path is not implemented yet, use file artifacts as a temporary fallback.
- This change focuses on unified CLI/scaffolding and refactors; it does not require implementing new SQLite persistence for currently missing stage outputs.

| Stage | Mode expectation in this change | Primary external expensive I/O | SQLite input tables | SQLite output tables | File fallback (temporary) |
|---|---|---|---|---|---|
| `init` | SQLite implemented | none | none | schema creation/upgrade for all pipeline tables | none |
| `collect-article-urls` | SQLite implemented | fccentrum.nl category pages | none | `articles_raw` (URL seed rows, minimal placeholders) | optional URL list artifact |
| `fetch-articles` | SQLite implemented | fccentrum.nl article pages | `articles_raw` (rows pending fetch) | `articles_raw` (HTML + `video_id` + status updates), `article_text_extractions`, `article_text_contents` | optional raw/debug payload artifacts |
| `acquire-audio` | SQLite implemented | YouTube via `yt-dlp` | `articles_raw` (`video_id`), existing `article_audio_sources` for dedupe | `article_audio_sources` | optional audio blob export artifact |
| `transcribe-audio` | SQLite implemented | Murmel API | `article_audio_sources` | `article_audio_transcriptions` | optional transcription JSON artifact |
| `extract-spots` | Hybrid: current SQLite record + file artifacts | Gemma API | `article_audio_transcriptions`, `article_audio_sources`, `articles_raw`, `article_text_contents` | `article_spot_extractions` (raw/parsed model output record) | stage artifacts for prompts/responses/intermediate outputs |
| `geocode-spots` | File-first scaffold (SQLite target deferred) | Google Places Text Search API | spot candidates from prior stage (SQLite-derived or file) | target table (`spots` and/or staging cache) **deferred in this change** | geocode result artifacts used as handoff |
| `export-data` | SQLite implemented | none (local file write only) | `spots`, `authors`, `articles`, `article_spots` | no DB write (writes static files only) | exported static files |

Notes:
- Backwards compatibility is not required; table contracts can evolve as long as full current functionality is preserved.
- For unimplemented SQLite outputs (currently geocode integration), file artifacts are the accepted temporary contract in this change.
- Invoking an unsupported stage/mode combination MUST return explicit error (non-zero exit), never silent success.
- Final state still removes legacy CLI entrypoints; temporary file fallback does not imply keeping legacy commands.

### Stage Mode Support Matrix (explicit)

| Stage | `--io sqlite` | `--io file` |
|---|---|---|
| `init` | Supported | Not supported (error) |
| `collect-article-urls` | Supported | Supported |
| `fetch-articles` | Supported | Supported |
| `acquire-audio` | Supported | Supported |
| `transcribe-audio` | Supported | Supported |
| `extract-spots` | Supported | Supported |
| `geocode-spots` | **Not supported yet (error)** | Supported |
| `export-data` | Supported | Supported (adapter/fallback as implemented) |

## Risks / Trade-offs

- **[Risk] Incomplete stage parity at first scaffold pass** → Mitigation: define strict stage contracts + explicit TODO markers + acceptance tasks per stage.
- **[Risk] File mode format churn across iterations** → Mitigation: version artifact payloads (`schemaVersion`) and lock required fields in specs.
- **[Risk] Database reset can disrupt local workflows** → Mitigation: provide backup/export before reset and clear migration docs.
- **[Risk] More subcommands increases surface area** → Mitigation: shared command framework, shared flags, centralized validation.

## Migration Plan

1. Add `scrape` root CLI with subcommand scaffolding and shared I/O/options layer.
2. Wire each stage to current internal logic in SQLite mode (parity-first adapters).
3. Add file mode read/write adapters per stage with default `data/stages/<stage>/` paths.
4. Introduce schema updates required by stage contracts; include reset path and docs.
5. (Optional during development only) keep temporary adapters while wiring parity.
6. Validate parity by running a single-article end-to-end pipeline via new subcommands (not full-corpus run).
7. Add `scrape init` preflight env-var validation and clear diagnostics for missing required API keys.
8. Add explicit unsupported-mode errors for stage/mode combinations not implemented (especially `geocode-spots --io sqlite`).
9. Remove legacy entrypoints/binaries and legacy-only code paths before marking change complete.

Rollback: use Git history/branch rollback if needed; do not keep permanent legacy command paths in the completed change.

## E2E Validation Commands (single article only, SQLite-first)

Policy:
- E2E validation uses one explicit article URL only.
- SQLite-first for every stage that already has SQLite behavior.
- Use file artifacts only where SQLite output is not implemented yet (currently geocode stage handoff).
- No automatic retries for expensive API stages.
- If an expensive stage fails (`acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`), stop and request explicit user approval before re-running.

Expected validation flow:

```bash
# 0) Preflight + fresh DB for deterministic validation
cd scraper
go run ./cmd/scrape init --db-path ../data/spots.db --reset
# init must fail fast if required API env vars are missing

# 1) Seed exactly one article (SQLite)
go run ./cmd/scrape collect-article-urls --io sqlite --db-path ../data/spots.db --article-url "<FCCENTRUM_ARTICLE_URL>"

# 2) Fetch article + store text extraction (SQLite)
go run ./cmd/scrape fetch-articles --io sqlite --db-path ../data/spots.db

# 3) Acquire audio blob (SQLite, expensive: YouTube)
go run ./cmd/scrape acquire-audio --io sqlite --db-path ../data/spots.db

# 4) Transcribe audio (SQLite, expensive: Murmel)
go run ./cmd/scrape transcribe-audio --io sqlite --db-path ../data/spots.db --language nl

# 5) Extract spots (SQLite + artifact outputs, expensive: Gemma)
go run ./cmd/scrape extract-spots --io sqlite --db-path ../data/spots.db

# 6) Geocode spots (file fallback handoff for now, expensive: Google Places)
# input should be explicit artifact path produced by prior stage contract
go run ./cmd/scrape geocode-spots --io file --in ../data/stages/extract-spots/<identity>__extract-spots__candidates.json

# 7) Export static data from SQLite (smoke test)
go run ./cmd/scrape export-data --io sqlite --db-path ../data/spots.db --out ../viz/public/data/spots.json
# Note: in this scaffold phase we do NOT expect meaningful spot output yet because
# upstream stages do not currently populate final export tables end-to-end.
# Expected successful smoke-test output may be:
# {"authors": null, "spots": null}
```

Notes:
- Command names/flags are the target interface for this change and may be scaffolded before full implementation.
- Stage 6 explicitly demonstrates temporary file-mode fallback due to deferred SQLite geocode writes.
- When debugging E2E failures, prefer local inspection and SQL/file diagnostics over re-calling external APIs.

## Open Questions

- None for current scope.
