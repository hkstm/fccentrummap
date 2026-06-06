## Why

The current `spots.json` export exposes very specific Google Places primary type display names, which are too granular for frontend category display and filtering. Introducing an export-time category mapping lets the database retain raw geocoding detail while publishing a stable, user-friendly category name contract.

## What Changes

- **BREAKING**: Replace the exported spot field for primary type display name with `categoryName` in `spots.json`.
- Add an export mapping step that converts stored `primary_type_display_name` values from the database into broader category names.
- Ensure spots remain exportable when no primary type display name is stored or no mapping is available, using a defined fallback/null behavior.
- Keep the SQLite schema storage unchanged so raw Google Places primary type display names remain available internally.

## Capabilities

### New Capabilities

### Modified Capabilities
- `scraping-data-json-export`: Exported spot JSON shall publish mapped category names instead of raw primary type display names.
- `static-data`: The frontend static data contract shall consume `categoryName` on spot records instead of the raw primary type display name field.

## Impact

- Affects the Go JSON export code that reads Gemini geocode rows and serializes spot records.
- Affects `viz/public/data/spots.json` structure and any frontend/types/tests that reference the old primary type display name field.
- Does not require database schema changes; `gemini_direct_spot_google_geocodes.primary_type_display_name` remains the source input for export mapping.
