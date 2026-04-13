## Why

The scraper change (`scrape-spots-to-sqlite`) produces a SQLite database of geocoded Amsterdam spots with author attribution, but there's no way to explore or visualize this data. A map-based frontend lets users see all spots at a glance, filter by author, and discover places spatially — which is the entire point of collecting the data.

## What Changes

- Add a Next.js (TypeScript) web application as the frontend for the spots database
- Render all spots on an interactive Google Maps map using the Advanced Marker API
- Custom markers styled as the Amsterdam "X" (andreaskruis) with randomly assigned colors from a distinct palette
- Author filter UI that lets users show/hide spots by author
- Build-time data extraction: read the SQLite database at build time and emit a static JSON file, enabling fully static deployment (no server required)

## Capabilities

### New Capabilities
- `map-view`: Interactive Google Maps integration — loading the map, rendering spots as Advanced Markers, and handling the marker lifecycle
- `custom-markers`: Amsterdam "X" styled markers using the Advanced Marker API with a color palette for visual distinction
- `author-filter`: UI control for filtering displayed spots by author, updating the map in real time
- `static-data`: Go command (`scraper/cmd/export`) that reads the SQLite database and emits a static JSON file containing all spots and authors, reusing existing Go/SQLite dependencies with no native Node modules required

### Modified Capabilities
None — no existing specs to modify.

## Impact

- **New dependency**: Next.js, React, TypeScript, `@googlemaps/js-api-loader` (or `@vis.gl/react-google-maps`)
- **No new Go dependencies**: The data export reuses `modernc.org/sqlite` already present from the scraper
- **Google Maps API**: Requires a Maps JavaScript API key with the Maps and Advanced Markers APIs enabled (may use the same project as the Geocoding API key from the scraper)
- **Database**: Read-only access to `data/spots.db` at build time only — no runtime dependency
- **Deployment**: Fully static export (`next export`) — deployable to GitHub Pages, Netlify, S3, or any static host
- **Project structure**: Adds a frontend app under `viz/` alongside the Go scraper module under `scraper/`
