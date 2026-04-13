## Context

The `scrape-spots-to-sqlite` change produces a SQLite database with ~150 geocoded Amsterdam spots linked to ~40-50 authors. This change adds a Next.js (TypeScript) frontend that visualizes these spots on Google Maps. The dataset is small and changes infrequently (only when the scraper is re-run), making it ideal for a fully static site where data is embedded at build time.

The scraper is a Go project. The frontend lives alongside it in the same repository but is an independent Next.js application with no runtime coupling to the Go code — they share only the SQLite database file at build time.

## Goals / Non-Goals

**Goals:**
- Render all spots on a Google Maps map using the Advanced Marker API
- Display markers as Amsterdam "X" (andreaskruis) icons with distinct colors from a palette
- Allow users to filter spots by author
- Deploy as a fully static site (no server runtime)
- Keep the frontend self-contained — `go run ./cmd/export -db ../data/spots.db -out ../viz/public/data/spots.json` + `npm run build` produces a deployable artifact with no native Node dependencies

**Non-Goals:**
- Server-side rendering or API routes — the entire site is statically exported
- Editing or writing to the database from the frontend
- Mobile-native app or PWA capabilities
- Search, sorting, or any filtering beyond author selection
- Info windows or detail panels for spots (can be added later)

## Decisions

### 1. `@vis.gl/react-google-maps` for Maps integration
**Choice:** Use the `@vis.gl/react-google-maps` library for Google Maps integration.
**Rationale:** This is the community-standard React wrapper for the Maps JavaScript API, maintained under the vis.gl umbrella. It provides `<APIProvider>`, `<Map>`, and `<AdvancedMarker>` components with full TypeScript support. The `<AdvancedMarker>` component accepts arbitrary React children as marker content, which is exactly what we need for custom Amsterdam "X" SVG markers.
**Alternatives considered:** `@googlemaps/js-api-loader` with raw DOM manipulation — more boilerplate and no React integration; `google-map-react` — older library, doesn't support Advanced Markers natively.

### 2. Custom markers via inline SVG React components
**Choice:** Render the Amsterdam "X" (andreaskruis) as an inline SVG passed as a child of `<AdvancedMarker>`. Each author gets a deterministic color from a predefined palette.
**Rationale:** The Advanced Marker API supports arbitrary HTML/React content as the marker element. An SVG andreaskruis is simple geometry (three overlapping saltire shapes) and renders crisply at any size. Using a React component means colors can be passed as props. Deterministic color assignment (based on author index or name hash) ensures consistent colors across page loads.
**Alternatives considered:** PNG marker images — poor scaling, harder to recolor dynamically; `google.maps.marker.PinElement` with custom glyphs — limited styling control compared to full SVG.

### 3. Color palette for author distinction
**Choice:** Define a fixed palette of 12-16 visually distinct colors. Assign colors to authors deterministically (e.g., sorted author list → palette index mod palette length).
**Rationale:** A curated palette ensures colors are distinguishable from each other and visible against the map. Deterministic assignment (not random at render time) means the same author always gets the same color. With ~40-50 authors the palette will wrap, but spatially distant authors sharing a color is acceptable since the filter UI disambiguates.
**Alternatives considered:** Fully random colors — risk of clashing or indistinguishable hues; per-author stored colors in DB — over-engineering for this use case.

### 4. Static data via Go exporter
**Choice:** A small Go command (`scraper/cmd/export/main.go`) reads the SQLite database using `modernc.org/sqlite` and writes `viz/public/data/spots.json`. This runs as a pre-build step before the frontend build.
**Rationale:** The dataset is small (~150 spots, ~50 authors) and changes only when the scraper is re-run. Embedding the data as a static JSON file means zero runtime server dependency. The frontend fetches this JSON file at page load. Using Go for the export reuses the existing `modernc.org/sqlite` dependency (pure Go, no CGo) already in the project from the scraper, and keeps the Node.js frontend completely free of native dependencies. The build step is: `cd scraper && go run ./cmd/export -db ../data/spots.db -out ../viz/public/data/spots.json`.
**Alternatives considered:** Node.js script with `better-sqlite3` — introduces a native C dependency (node-gyp) into the frontend build, causing potential compilation issues across environments; `sql.js` (Wasm) — avoids native deps but adds a large Wasm binary as a dev dependency for a trivial task; `getStaticProps` reading SQLite directly — same native module bundling issues.

### 5. Next.js with static export (`output: 'export'`)
**Choice:** Use the Next.js App Router with `output: 'export'` in `next.config.ts` to produce a fully static site.
**Rationale:** The app is a single-page map viewer with no dynamic routes or server-side logic. Static export produces an `out/` directory with plain HTML/CSS/JS deployable to any static host (GitHub Pages, Netlify, S3). The App Router is the modern Next.js standard and supports static export fully for this use case.
**Alternatives considered:** Pages Router — still supported but App Router is the recommended path; Vite + React — viable but Next.js gives us a more opinionated setup with less configuration.

### 6. Author filter as a sidebar/panel with checkboxes
**Choice:** A collapsible panel (sidebar or overlay) listing all authors with checkboxes. All authors are selected by default. Unchecking an author hides their spots from the map. Include a "select all / deselect all" toggle.
**Rationale:** Checkboxes are the simplest UX for multi-select filtering. A panel keeps the map uncluttered while remaining accessible. The author list is derived from the same static JSON — no additional data fetching needed.
**Alternatives considered:** Dropdown multi-select — harder to scan with 40+ authors; map-layer toggles — adds complexity without UX benefit.

### 7. Project structure
**Choice:** Place the frontend app under `viz/` and keep the Go module under `scraper/`.
**Rationale:** The repo now has explicit subsystem boundaries. `scraper/` owns the data pipeline, while `viz/` is the logical home for the frontend deliverable.

```
├── scraper/
│   └── cmd/export/main.go  # SQLite -> static JSON export
├── viz/
│   ├── app/                # Next.js App Router pages
│   ├── components/         # React components
│   ├── lib/                # Shared utilities
│   ├── public/
│   │   └── data/
│   │       └── spots.json  # Generated at build time (gitignored)
│   ├── next.config.ts
│   ├── package.json
│   └── tsconfig.json
└── data/
    └── spots.db            # Generated scraper database
```

### 8. Google Maps Map ID
**Choice:** Require a Cloud-based Map ID configured via environment variable (`NEXT_PUBLIC_GOOGLE_MAPS_MAP_ID`).
**Rationale:** Advanced Markers require a Map ID linked to a cloud-based map style. This is a Google Maps platform requirement, not a design choice. The Map ID is safe to embed client-side (it's not a secret — access is controlled by API key restrictions). The API key is also provided via `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY`.

## Risks / Trade-offs

- **[Advanced Markers require Map ID]** → A cloud-based map style must be created in Google Cloud Console before the app works. Mitigation: document the setup steps clearly.
- **[Color palette wrapping]** → With ~40-50 authors and 12-16 palette colors, multiple authors share colors. Mitigation: acceptable because the filter UI disambiguates; spots from different authors rarely overlap spatially.
- **[Stale data]** → The static JSON is only updated when the build runs. Mitigation: the underlying data only changes when the scraper is re-run, which is infrequent. Re-deploy after scraping.
- **[Go toolchain required for build]** → The data export step requires Go installed on the build machine. Mitigation: Go is already required for the scraper, so this adds no new build dependency. The exported JSON can also be committed to the repo if a Go-free frontend build is needed.
- **[API key exposure]** → The Maps API key is embedded in client-side JavaScript. Mitigation: restrict the key to the Maps JavaScript API and limit by HTTP referrer in Google Cloud Console.
