## Why

We need a fast, standalone way to resolve a place name into latitude/longitude using Google Places Text Search. Building this as an isolated function plus a tiny CLI wrapper enables quick debugging and validation before any broader pipeline integration.

## What Changes

- Add a focused geocoding function that takes a place name string as input and returns latitude/longitude output.
- Use Google Places Text Search as the backing API for the lookup.
- Add a small CLI wrapper to call the function directly from terminal and print JSON output by default for debugging and scripting.
- Define CLI success JSON as identity-first: `query`, `name`, `placeId`, and `mapsUrl` (derived from query + placeId); latitude/longitude are intentionally not required in CLI output.
- Apply a hard Google Places Text Search `locationRestriction.rectangle` bounded by low `(52.274525, 4.711585)` and high `(52.461764, 5.073559)` so results must fall within the target region.
- Define clear error behavior for missing API key, empty input, and no-match API responses (including no results inside the restriction).

## Capabilities

### New Capabilities
- `google-places-text-search-geocoding`: Resolve a single place-name query to latitude/longitude via Google Places Text Search through a reusable Go function and CLI debug entrypoint, with JSON-first output and enforced rectangular location restriction.

### Modified Capabilities
- None.

## Impact

- Affected code: `scraper/internal/geocoder` and a new lightweight CLI command for manual execution/debugging.
- External API/dependency impact: Google Places Text Search request/response handling and API-key configuration.
- Persisted/export pipeline contracts are unchanged in this change; coordinates remain guaranteed at the reusable function layer (lat/lng), while CLI output is identity-first for debugging.
- No SQLite schema or export contract changes in this change.
