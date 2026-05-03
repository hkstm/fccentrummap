## Context

The current pipeline can scrape articles, transcribe audio, and extract candidate spot mentions with timestamps, but coordinate resolution is not consistently integrated into persisted/exported outputs. The repository already has a `geocoder` package and SQLite-backed export flow, so this change should introduce a deterministic geocoding stage that resolves extracted place names to latitude/longitude using Google Places Text Search and makes those coordinates available to downstream map data.

Key constraints:
- Keep existing extraction flow behavior stable while adding geocoding as an additional step.
- Respect Google Places API quota/error modes and avoid non-deterministic persistence behavior.
- Preserve inspectability: unresolved and ambiguous matches must remain observable for later review.

## Goals / Non-Goals

**Goals:**
- Add a Google Places Text Search integration that resolves extracted place names to geospatial coordinates (lat/lng).
- Persist geocoding outcomes in SQLite in a way that can be re-read and exported consistently.
- Define deterministic behavior for no-match, multi-match, and API-failure cases.
- Include resolved coordinates in exported static data consumed by visualization.

**Non-Goals:**
- Building a full place-disambiguation UI.
- Implementing reverse geocoding or nearby search.
- Solving all historical data quality issues in one pass.
- Replacing transcript extraction logic itself.

## Decisions

1. **Use Google Places Text Search as the canonical resolver for extracted spot names.**
   - **Why:** It aligns with the required API and provides place candidates plus geometry in one request flow.
   - **Alternative considered:** Existing generic geocoding provider wrappers only; rejected because this change explicitly targets Places Text Search semantics and response model.

2. **Add explicit geocoding result status to persisted extraction-linked records.**
   - **Why:** We need deterministic outcomes for `resolved`, `unresolved`, and `error` to support retries, audits, and exports.
   - **Alternative considered:** Persist only successful coordinates; rejected because silent drops would hide pipeline behavior and make debugging hard.

3. **Use a conservative first-result selection policy with stored raw candidate metadata.**
   - **Why:** Keeps implementation simple and deterministic while preserving enough detail for future disambiguation improvements.
   - **Alternative considered:** Scoring across multiple candidates with fuzzy matching; deferred to future change due to complexity and unclear acceptance criteria.

4. **Integrate geocoding output into existing export path rather than adding a separate export artifact.**
   - **Why:** Consumers already read `spots.json`; extending this contract minimizes integration overhead.
   - **Alternative considered:** New parallel geodata file; rejected to avoid dual-source coordination.

5. **Gate API configuration via explicit env/flag validation before geocoding runs.**
   - **Why:** Fail-fast validation gives clearer operator feedback and avoids partial writes under missing credentials.
   - **Alternative considered:** Best-effort execution with deferred errors; rejected because it produces inconsistent run results.

## Risks / Trade-offs

- **[Risk] Place-name ambiguity can map to the wrong location** → Mitigation: store candidate metadata/status and keep deterministic first-pass policy explicit in specs.
- **[Risk] API quota limits or transient failures interrupt runs** → Mitigation: classify failures distinctly from unresolved results and support safe reruns.
- **[Risk] Schema changes affect existing data/export assumptions** → Mitigation: constrain migration to additive fields/tables where possible and keep backward-compatible export shape where feasible.
- **[Risk] Tight coupling between extraction and geocoding stages** → Mitigation: keep a clear interface boundary so geocoding can be executed/retried independently when needed.

## Migration Plan

1. Add/adjust SQLite schema elements required to persist geocoding outcomes tied to extracted places.
2. Implement repository read/write methods for geocoding status, selected result, and coordinate fields.
3. Integrate Google Places Text Search client logic and configuration validation in scraper pipeline command paths.
4. Update export mapping to include resolved coordinates in static output.
5. Backfill or rerun extraction/geocoding for selected records as needed in development data.

Rollback:
- Revert code paths to pre-geocoding behavior.
- Keep migration safety notes and backup strategy for any destructive schema resets.

## Open Questions

- Should first-result selection require a minimum confidence/signal threshold before marking `resolved`?
- Do we need per-query caching (by normalized place text) in this change, or defer until quota pressure appears?
- Should unresolved/error outcomes be exported to frontend diagnostics, or remain DB-only for now?
