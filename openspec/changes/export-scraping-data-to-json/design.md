## Context

Scraped Amsterdam spot data is currently stored in SQLite and is useful for internal processing, but the target delivery model is a static website. The frontend should be able to load a bundled JSON file at request time without requiring a runtime database dependency. The export must therefore provide a stable, deterministic shape that can be generated from current relational tables and consumed by static assets.

## Goals / Non-Goals

**Goals:**
- Produce a static JSON dump from SQLite as part of the scraping/export workflow.
- Define a stable payload contract with top-level `spots` and `presenters` arrays.
- Ensure each spot includes `placeId`, `spotName`, `presenterName`, and `youtubeLink`.
- Support configurable output path/filename via CLI or config.
- Keep output deterministic to reduce noisy diffs in generated artifacts.

**Non-Goals:**
- Replacing SQLite as the system of record.
- Building a dynamic API server for this data.
- Introducing frontend-side schema transformation logic beyond direct payload consumption.
- Solving broader versioned CDN cache invalidation in this change.
- Including payload version metadata (`schemaVersion`, `generatedAt`) in v1 output.
- Including additional identifiers beyond `placeId` (e.g., internal numeric IDs) in v1 output.

## Decisions

1. **Export from curated read-model queries rather than raw table dumps**  
   - **Why:** Table schemas are internal and may contain join tables or fields not appropriate for frontend contracts. A curated read-model lets us keep relational flexibility while shipping a stable external shape.
   - **Alternative considered:** Direct table-to-JSON serialization. Rejected because it leaks storage internals and creates fragile coupling.

2. **Single JSON document with explicit top-level collections (`spots`, `presenters`)**  
   - **Why:** Aligns with static bundle usage and keeps client loading simple (one fetch, one parse).
   - **Contract shape (v1):**
     ```json
     {
       "spots": [
         {
           "placeId": "someId",
           "spotName": "Some spot",
           "presenterName": "Ray Fuego",
           "youtubeLink": "https://www.youtube.com/watch?v=some-video-id"
         }
       ],
       "presenters": [
         {
           "presenterName": "Ray Fuego"
         }
       ]
     }
     ```
   - **Presenter collection semantics:** `presenters` is a deduplicated list derived from spot associations, with one object per normalized presenter name in v1.
   - **Alternative considered:** Multiple files per entity type. Rejected for now to avoid manifest complexity and ordering concerns.

3. **Deterministic ordering in exported arrays**  
   - **Why:** Stable ordering improves reproducibility, reviewability, and cache behavior.
   - **Approach:** Sort spots by a stable key (e.g., `placeId`) and presenters by `presenterName`.
   - **Alternative considered:** DB/default iteration order. Rejected due to nondeterministic output across runs.

4. **Presenter names are exported as-is from the database (no normalization in v1)**  
   - **Why:** The immediate goal is faithful static export for site bundling, not data cleaning. Keeping values unchanged avoids accidental transformation and keeps behavior transparent.
   - **Alternative considered:** Lightweight normalization (trim/case-fold) or canonical mapping tables. Rejected for v1 to avoid introducing opinionated data mutation.

5. **Graceful handling of partial/empty data**  
   - **Why:** Scraping runs can legitimately produce sparse datasets. Export should still emit valid JSON with empty arrays rather than fail hard when no rows are present.
   - **Alternative considered:** Treat empty result sets as errors. Rejected because it would block static build workflows unnecessarily.

6. **Use standard library JSON encoding and file I/O**  
   - **Why:** No new dependency is needed; Go stdlib is sufficient and keeps maintenance low.
   - **Alternative considered:** Third-party JSON libs for speed/features. Rejected as unnecessary for expected dataset size and requirements.

## Risks / Trade-offs

- **[Risk] Divergence between DB schema changes and export query mapping** → **Mitigation:** Centralize export mapping logic and add tests validating required fields are always present.
- **[Risk] Duplicate/inconsistent presenter naming in source data** → **Mitigation:** Export values as-is in v1; treat data cleanup/dedup strategy as a future, explicit data-quality change.
- **[Risk] Large exports may increase build artifact size** → **Mitigation:** Keep schema minimal and evaluate optional compression at static hosting layer.
- **[Trade-off] Curated export adds maintenance surface** → **Mitigation:** Treat JSON contract as explicit capability spec and validate in tests.

## Migration Plan

1. Implement read-model query layer for export payload fields.
2. Implement JSON writer with deterministic ordering and atomic file write behavior.
3. Add CLI/config flag(s) for output path and invocation semantics.
4. Add tests for schema shape, empty dataset output, and deterministic ordering.
5. Expose export via CLI option (with configurable output path) and keep it optional.
6. Rollback strategy: disable or stop using the export CLI option while retaining existing scraping/storage flow.

## Open Questions

