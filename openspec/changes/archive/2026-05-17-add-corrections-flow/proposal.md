## Why

The scraped/exported spot data occasionally needs manual corrections for display names, Google place IDs, and YouTube timestamps after automated extraction and geocoding. Today those fixes either require mutating source tables or hand-editing exported JSON, which makes corrections hard to audit and easy to lose on the next export.

## What Changes

- Add a correction workflow that lets maintainers create/update per-spot overrides keyed by a stable spot identifier from map share state.
- Support optional correction fields for spot name, Google `placeId`, and timestamp; omitted/empty inputs keep the original scraped value.
- When `placeId` is corrected, refresh the associated latitude/longitude from Google Places/Maps data so the exported marker moves with the corrected place.
- Store corrections in a separate SQLite table instead of mutating extracted spot/source rows.
- Apply corrections during JSON export so frontend output reflects overrides while the original extraction data remains available.
- Preserve existing exported JSON shape unless a stable correction/share identifier must be included for maintainers to target corrections.

## Capabilities

### New Capabilities
- `spot-corrections`: Maintainer workflow for recording spot field overrides, resolving corrected place coordinates, and applying corrections during export.

### Modified Capabilities
- `sqlite-storage`: Add durable correction storage and repository operations without mutating source extraction/geocode tables.
- `scraping-data-json-export`: Apply stored corrections when producing export data and fail clearly when corrected place coordinates cannot be resolved.
- `static-site-data-dump`: Expose a stable spot correction/share identifier while keeping existing spot fields compatible.
- `map-view`: Use stable spot identifiers in spot-specific share state.

## Impact

- Affected scraper CLI/repository/export code: SQLite schema initialization, correction upsert/query logic, export data assembly, and Google place coordinate lookup for corrected `placeId` values.
- Affected frontend code: share-state and type/data handling for a stable spot identifier.
- Affected data: new SQLite correction table; generated `viz/public/data/spots.json` may reflect corrected names, place IDs, coordinates, and timestamps.
- No new public backend API or runtime service is expected; this remains a local maintainer workflow plus static JSON export.
