## Why

We have geocoded Amsterdam spots with presenter attribution in a JSON dataset, but there's no way to explore or visualize this data. A map-based frontend lets users see all spots at a glance, filter by presenter, and discover places spatially.

## What Changes

- Add a Next.js (TypeScript) web application as the frontend for the spots dataset
- Render all spots on an interactive Google Maps map using the Advanced Marker API
- Custom markers styled as the Amsterdam "X" (andreaskruis) with deterministically assigned colors from a distinct palette
- Author filter UI that lets users show/hide spots by author
- Marker interaction that opens each spot’s `youtubeLink` (including timestamp) when selected
- Apply the FC Centrum visual system from `docs/fccentrum-styleguide.md` to UI surfaces (tokens, typography, component states, responsive behavior, accessibility baseline)
- Keep marker colors as a separate map-encoding palette (not styleguide-defined), but constrained for visual distinction and accessibility contrast
- Consume a static JSON input contract (`spots` + `presenters`) so deployment remains fully static (no server required), regardless of how the JSON is produced

## Capabilities

### New Capabilities
- `map-view`: Interactive Google Maps integration — loading the map, rendering spots as Advanced Markers, handling marker lifecycle, and marker click-through to timestamped video links
- `custom-markers`: Amsterdam "X" styled markers using the Advanced Marker API with a color palette for visual distinction
- `author-filter`: UI control for filtering displayed spots by presenter/author, updating the map in real time
- `static-data`: Frontend consumes a static JSON artifact (`/data/spots.json`) with spots and presenters, independent from the upstream generation pipeline

### Modified Capabilities
None — no existing specs to modify.

## Impact

- **New dependency**: Next.js, React, TypeScript, `@googlemaps/js-api-loader` (or `@vis.gl/react-google-maps`)
- **Data pipeline decoupling**: frontend depends only on the JSON contract, not on exporter implementation details
- **Google Maps API**: Requires a Maps JavaScript API key with the Maps and Advanced Markers APIs enabled (may use the same project as the Geocoding API key from the scraper)
- **Data input**: Read-only static JSON input (`/data/spots.json`) — no runtime backend dependency
- **Deployment**: Fully static export (`next export`) — deployable to GitHub Pages, Netlify, S3, or any static host
- **Project structure**: Adds a frontend app under `viz/` with static data input under `viz/public/data/`
- **Design alignment**: `render-spots-map` must implement the FC Centrum styleguide at `docs/fccentrum-styleguide.md` for UI surfaces and log any intentional deviations
- **Marker color policy**: marker author-colors are intentionally outside the styleguide and use a deterministic, high-contrast distinction palette
