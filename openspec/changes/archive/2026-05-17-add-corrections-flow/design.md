## Context

The scraper pipeline stores extracted spot mentions, resolved Google place identities, and article/video metadata in SQLite, then exports static JSON for the frontend. Some exported spots need manual correction after review: a cleaner display name, a better Google `placeId`, or a corrected YouTube timestamp. These fixes should survive re-exporting without overwriting the original extraction/geocoding records.

The frontend already supports spot-specific share state through a URL query parameter, but the current key is derived from visible fields. A correction workflow needs a stable identifier that remains usable even when the display name or place ID is overridden.

## Goals / Non-Goals

**Goals:**
- Add a maintainer-only correction flow keyed by a stable spot identifier from map share state.
- Store corrections in a separate SQLite table with nullable override fields.
- Let blank correction inputs preserve existing source values.
- Resolve corrected `placeId` values to latitude/longitude before export uses them.
- Apply corrections at export time while preserving the existing extraction/geocode source records.
- Cover repository, CLI, export, and frontend share-state behavior with focused tests.

**Non-Goals:**
- Building a public correction API or authenticated admin web UI.
- Mutating extracted spot mentions, article rows, or original geocode rows when corrections are applied.
- Supporting arbitrary field corrections beyond spot display name, Google place ID/location, and YouTube timestamp in this change.
- Adding user-facing correction submissions.

## Decisions

1. **Use a stable exported `spotId` as the correction key**
   - **Decision:** Export each spot with a stable `spotId` derived from immutable source identity, preferably the underlying article/spot relationship ID or a deterministic repository-level key that does not include corrected fields.
   - **Rationale:** A key built from `spotName` or `placeId` becomes unstable exactly when those fields are corrected. The frontend can put `spotId` in share state, and the CLI can target the same value.
   - **Alternatives considered:**
     - Use the current share-state key: convenient, but it changes when corrected fields change.
     - Use `placeId`: not unique across presenters/articles and changes when correcting the place.

2. **Store corrections separately from source data**
   - **Decision:** Add a `spot_corrections` table keyed by `spot_id` with nullable `spot_name`, `place_id`, `latitude`, `longitude`, and timestamp override fields plus timestamps for auditability.
   - **Rationale:** Source extraction/geocoding remains reproducible, and corrections are explicit and idempotent.
   - **Alternatives considered:**
     - Update `article_spots` / geocode rows directly: simpler export, but loses original data and makes future reprocessing harder to reason about.
     - Patch generated JSON manually: quick but non-durable and overwritten on export.

3. **Use an interactive prompt-only CLI for the first implementation**
   - **Decision:** The correction command SHALL be interactive in the first implementation: maintainers run it with a stable spot identifier, then the command displays current effective/source values and prompts for optional updated name, place ID, and timestamped YouTube URL. Empty input leaves that field unchanged. Non-interactive correction flags are out of scope for this change.
   - **Rationale:** This matches the intended manual review workflow and minimizes accidental clearing of valid data while keeping the first implementation focused.
   - **Alternatives considered:**
     - Require every field every time: noisy and error-prone.
     - Add non-interactive flags immediately: useful for scripting, but not needed for the current maintainer workflow and can be added later.

4. **Resolve corrected `placeId` to coordinates before saving or exporting**
   - **Decision:** When a correction supplies a new `placeId`, resolve coordinates via a Google place-id lookup and persist the resulting latitude/longitude with the correction. Export fails if a corrected `placeId` lacks coordinates.
   - **Rationale:** The map marker must move with the corrected Google place. Persisting coordinates avoids repeating API calls on every export.
   - **Alternatives considered:**
     - Reuse old coordinates: incorrect for changed places.
     - Resolve coordinates lazily on each export: simpler writes, but makes export network-dependent and slower.

5. **Apply corrections only in the export assembly layer**
   - **Decision:** Repository export queries SHALL return source values plus correction values or already-effective values, and the export service SHALL emit corrected values in the same JSON structure.
   - **Rationale:** The frontend remains simple and sees only the effective static data.
   - **Alternatives considered:**
     - Teach the frontend to apply corrections: leaks maintenance details into runtime UI and requires shipping extra data.

6. **Timestamp corrections accept only full timestamped YouTube URLs**
   - **Decision:** The timestamp prompt SHALL accept only full `youtube.com` or `youtu.be` URLs that contain a timestamp. The system extracts the timestamp from the URL and applies that timestamp to the existing source video URL during export.
   - **Rationale:** Maintainers can copy/paste the exact timestamped YouTube URL they verified, while the correction remains scoped to timestamp correction rather than arbitrary video replacement.
   - **Alternatives considered:**
     - Accept raw seconds, `MM:SS`, or `HH:MM:SS`: concise but easier to mistype and detached from the reviewed YouTube context.
     - Store the pasted URL as the full exported URL: flexible, but changes the correction from timestamp-only to video/link replacement.

## Risks / Trade-offs

- **[Risk] Stable ID choice may be hard to derive for existing rows** → **Mitigation:** Prefer existing join-table primary keys; otherwise introduce a deterministic source identity in the export query and tests proving stability across correction changes.
- **[Risk] Google place-id coordinate lookup can fail or require a different API endpoint than text search** → **Mitigation:** Isolate place-id lookup behind a small interface with mocked tests and actionable errors.
- **[Trade-off] Persisted corrected coordinates can become stale if Google data changes** → **Mitigation:** Corrections are maintainer-managed; updating the place ID or re-saving the correction can refresh coordinates.
- **[Risk] Timestamp URL parsing can reject valid but uncommon YouTube URL shapes** → **Mitigation:** Support and test the common timestamped `youtube.com` and `youtu.be` forms first, and fail with clear guidance for unsupported or untimestamped URLs.

## Migration Plan

- Add the correction table idempotently during SQLite repository initialization.
- Existing databases start with no corrections and export unchanged data.
- Add correction creation/update command and repository tests before changing export behavior.
- Regenerate static JSON only after corrections are applied.
- Rollback is removing/ignoring correction rows; source extraction and geocode records remain intact.

## Open Questions

None.
