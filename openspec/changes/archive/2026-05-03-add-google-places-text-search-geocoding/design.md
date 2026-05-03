## Context

This change is intentionally scoped as a standalone utility, not a pipeline integration. We need a simple, testable function that accepts a place name and returns coordinates, plus a tiny CLI wrapper so developers can run lookups manually and inspect results during debugging.

Constraints:
- Keep implementation independent from existing extraction/storage/export flows.
- Use Google Places Text Search as the lookup mechanism.
- Enforce hard location restriction (not bias) using `locationRestriction.rectangle` with low `(52.274525, 4.711585)` and high `(52.461764, 5.073559)`.
- Provide clear operator feedback for invalid input and API failures.

## Goals / Non-Goals

**Goals:**
- Implement a reusable Go function with interface: input `placeName string` → output `(latitude, longitude)` (plus error).
- Integrate Google Places Text Search HTTP call and response parsing needed for the first matched result.
- Add a minimal CLI command that accepts a place name argument/flag and prints JSON output by default with resolved coordinates.
- Ensure deterministic failure behavior for empty query, missing API key, no results (including no matches within restriction), and HTTP/API errors.

**Non-Goals:**
- Integrating geocoding into transcript extraction or scraper pipelines.
- Persisting geocoding output to SQLite.
- Updating export/static frontend datasets.
- Implementing advanced ranking/disambiguation across multiple candidates.

## Decisions

1. **Expose a small dedicated geocoding function in `internal/geocoder`.**
   - **Why:** Keeps logic reusable by future integrations while staying easy to unit-test now.
   - **Alternative considered:** Put all logic inside a CLI command; rejected because it reduces reusability.

2. **Apply hard geographic filtering via `locationRestriction.rectangle` instead of location bias.**
   - **Why:** The requirement is to constrain results strictly to the target region, not just prefer it.
   - **Alternative considered:** `locationBias`; rejected because it may return results outside the intended area.

3. **Return first valid Places result as the resolved coordinate pair.**
   - **Why:** Provides deterministic, minimal behavior suitable for a debugging-oriented first step.
   - **Alternative considered:** Return all candidates; deferred to a later change if needed.

4. **Add a thin CLI wrapper command for manual lookup with JSON as default output format.**
   - **Why:** JSON is easier for scripting and future extensibility while still being human-inspectable.
   - **Alternative considered:** Plain `lat,lng` as default; rejected because it is less extensible for debugging metadata.

5. **Use fail-fast config and input validation.**
   - **Why:** Better developer ergonomics and clearer errors when API key/query is missing.
   - **Alternative considered:** Attempt request and rely on remote errors; rejected because local validation is clearer.

## Risks / Trade-offs

- **[Risk] First-result selection may be incorrect for ambiguous place names** → Mitigation: document behavior clearly and treat this as a baseline implementation.
- **[Risk] Restriction box may exclude expected edge-case results** → Mitigation: return explicit no-result-in-restriction errors and allow future configurable bounds.
- **[Risk] Quota/network failures reduce CLI reliability** → Mitigation: return actionable errors and preserve simple retry behavior.
- **[Risk] API response shape changes could break parsing** → Mitigation: keep parsing strict but minimal and cover with unit tests for representative payloads.

## Migration Plan

1. Implement/extend geocoder client function for text search lookup by place name.
2. Add unit tests for parsing and error paths.
3. Add a minimal CLI command (e.g., `cmd/geocode-place`) that calls the function and prints coordinates.
4. Document command usage in README/development docs.

Rollback:
- Remove the new CLI command and standalone function usage; no data migration or schema rollback required.

## Open Questions

- None at this stage; default output is JSON and location filtering uses fixed rectangle restriction bounds.
