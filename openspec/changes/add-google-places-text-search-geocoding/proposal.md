## Why

We can already extract spot names from article/transcript data, but those spots are not consistently resolved to reliable coordinates. Integrating Google Places Text Search now enables a deterministic way to produce latitude/longitude outputs for extracted places so downstream map visualization can use real geospatial points.

## What Changes

- Add a geocoding resolution step that queries Google Places Text Search for extracted spot names.
- Produce normalized geocoding output per spot candidate for pipeline/storage/export (maps/visualization), including at minimum `displayName`, `latitude`, `longitude`, `placeId`, and `mapsUrl`.
- Define handling for unresolved/ambiguous searches so pipeline output remains deterministic and inspectable.
- Keep CLI JSON/debug output identity-first (`query`, `name`, `placeId`, `mapsUrl`) where latitude/longitude may be omitted, while persisted/export outputs are where coordinates are guaranteed.
- Wire resolved coordinates into persisted pipeline outputs used by export/visualization flows.

## Capabilities

### New Capabilities
- `google-places-text-search-geocoding`: Resolve extracted place names via Google Places Text Search and return structured coordinates (lat/lng) with result metadata.

### Modified Capabilities
- `sqlite-storage`: Extend storage requirements so extracted/resolved spot geodata can be persisted and re-read reliably.
- `static-data`: Extend export requirements so generated frontend data includes coordinates sourced from geocoding resolution.

## Impact

- Affected code: `scraper/internal/geocoder`, extraction flow orchestration, repository read/write paths, and export pipeline.
- External API/dependency impact: Google Places Web Service Text Search request/response handling, API key/configuration, quota/error handling.
- Data impact: additional persisted geocoding fields and updated exported spot records with latitude/longitude.
