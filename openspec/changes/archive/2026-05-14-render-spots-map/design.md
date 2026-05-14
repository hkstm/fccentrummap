## Context

This change adds a Next.js (TypeScript) frontend that visualizes geocoded Amsterdam spots on Google Maps. The dataset is small and changes infrequently, making it ideal for a fully static site that reads a versioned JSON snapshot at build/runtime.

The frontend is intentionally data-source agnostic: it consumes the JSON contract only and does not depend on how that JSON is produced.

## Input data format (current exporter contract)

Current reference file: `data/export/spots.json`

```json
{
  "spots": [
    {
      "placeId": "string",
      "spotName": "string",
      "latitude": "number",
      "longitude": "number",
      "presenterName": "string",
      "youtubeLink": "string (url)"
    }
  ],
  "presenters": [
    {
      "presenterName": "string"
    }
  ]
}
```

Notes:
- The current sample contains one presenter, but the format already supports multiple presenters.
- `spots[*].presenterName` links each spot to a presenter entry.
- `spots[*].youtubeLink` is a deeplink target (including timestamp) and is used as the primary click-through action from the map marker.
- Frontend filtering and color assignment must assume many presenters (not single-presenter-only behavior).

## Goals / Non-Goals

**Goals:**
- Render all spots on a Google Maps map using the Advanced Marker API
- Default the initial map viewport to the Amsterdam area (matching existing geocoding bounds)
- Display markers as Amsterdam "X" (andreaskruis) icons with distinct colors from a palette
- Allow users to filter spots by author
- Allow users to click a spot marker and open the corresponding `youtubeLink` (including timestamp)
- Implement the FC Centrum design language defined in `docs/fccentrum-styleguide.md` for UI surfaces (tokens, typography hierarchy, component states, responsive rules, accessibility baseline)
- Use a separate deterministic marker-color palette for author encoding; palette is not prescribed by the styleguide and must prioritize distinction + accessibility contrast
- Deploy as a fully static site (no server runtime)
- Keep the frontend self-contained — as long as the JSON input contract is present, `npm run build` produces a deployable artifact

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
**Choice:** Render the Amsterdam "X" (andreaskruis) as an inline SVG passed as a child of `<AdvancedMarker>`. Each author gets a deterministic color from a predefined palette, and each marker is interactive.
**Rationale:** The Advanced Marker API supports arbitrary HTML/React content as the marker element. An SVG andreaskruis is simple geometry (three overlapping saltire shapes) and renders crisply at any size. Using a React component means colors can be passed as props. Deterministic color assignment (based on author index or name hash) ensures consistent colors across page loads.
**Interaction behavior:** Clicking (or keyboard-activating) a marker opens the spot's `youtubeLink` in a new tab, preserving any timestamp in the URL.
**Alternatives considered:** PNG marker images — poor scaling, harder to recolor dynamically; `google.maps.marker.PinElement` with custom glyphs — limited styling control compared to full SVG.

### 3. Color palette for author distinction
**Choice:** Define a fixed palette of 12-16 visually distinct colors. Assign colors to authors deterministically (e.g., sorted author list → palette index mod palette length).
**Rationale:** The FC Centrum styleguide does not define category/marker color sets for map encodings, so marker colors are specified separately in this change. A curated palette ensures colors are distinguishable from each other and visible against the map. Deterministic assignment (not random at render time) means the same author always gets the same color. With ~40-50 authors the palette will wrap, but spatially distant authors sharing a color is acceptable since the filter UI disambiguates.
**Constraints:** Maintain sufficient contrast and avoid near-indistinguishable hues; treat this as a data-encoding palette, not brand-accent usage.
**Alternatives considered:** Fully random colors — risk of clashing or indistinguishable hues; per-author stored colors in DB — over-engineering for this use case.

### 4. Static data contract (JSON-first)
**Choice:** Frontend reads a static JSON file (`/data/spots.json`) that follows the documented input contract.
**Rationale:** The dataset is small and mostly append/update over time. Using a static JSON artifact keeps deployment simple (no runtime backend) and cleanly decouples UI implementation from ingestion/export implementation.
**Alternatives considered:** Runtime API/backend fetch — unnecessary operational complexity for this use case.

### 5. Next.js with static export (`output: 'export'`)
**Choice:** Use the Next.js App Router with `output: 'export'` in `next.config.ts` to produce a fully static site.
**Rationale:** The app is a single-page map viewer with no dynamic routes or server-side logic. Static export produces an `out/` directory with plain HTML/CSS/JS deployable to any static host (GitHub Pages, Netlify, S3). The App Router is the modern Next.js standard and supports static export fully for this use case.
**Alternatives considered:** Pages Router — still supported but App Router is the recommended path; Vite + React — viable but Next.js gives us a more opinionated setup with less configuration.

### 6. Author filter as a sidebar/panel with checkboxes
**Choice:** A collapsible panel (sidebar or overlay) listing all authors with checkboxes. All authors are selected by default. Unchecking an author hides their spots from the map. Include a "select all / deselect all" toggle.
**Rationale:** Checkboxes are the simplest UX for multi-select filtering. A panel keeps the map uncluttered while remaining accessible. The author list is derived from the same static JSON — no additional data fetching needed.
**Alternatives considered:** Dropdown multi-select — harder to scan with 40+ authors; map-layer toggles — adds complexity without UX benefit.

### 7. Project structure
**Choice:** Place the frontend app under `viz/`, with data consumed via `viz/public/data/spots.json` (or equivalent static path).
**Rationale:** Keeps clear boundaries: the frontend owns rendering/filtering UX and depends only on the JSON contract.

```
├── viz/
│   ├── app/                # Next.js App Router pages
│   ├── components/         # React components
│   ├── lib/                # Shared utilities
│   ├── public/
│   │   └── data/
│   │       └── spots.json  # Static input artifact
│   ├── next.config.ts
│   ├── package.json
│   └── tsconfig.json
└── data/
    └── export/
        └── spots.json      # Source example / pipeline output
```

### 8. Google Maps Map ID + default Amsterdam viewport
**Choice:** Use the demo map ID (`DEMO_MAP_ID`) for this PoC change, and initialize the map to Amsterdam bounds aligned with the existing geocoding rectangle.
**Rationale:** Advanced Markers require a Map ID. For PoC scope, `DEMO_MAP_ID` keeps setup minimal while staying compatible with demo-key prototyping. Aligning the initial viewport to the geocoding rectangle gives users immediate context in the target city.

**Default viewport policy:**
- Initial map view should fit the Amsterdam bounding box currently used by geocoding:
  - low: `52.274525, 4.711585`
  - high: `52.461764, 5.073559`

**Environment key policy (PoC phase):**
- Use `DEMO_GOOGLE_MAPS_API_KEY` (Maps Demo Key) for map rendering during this change.
- Production key and custom Map ID strategy are intentionally deferred and will be defined when/if the project is productionized.

### 9. External documentation currency policy (Google Maps)
**Choice:** For every Google Maps-related implementation decision in this change, consult the latest official Google documentation from the internet before coding.
**Rationale:** Maps APIs, billing behavior, and feature requirements can change. Pulling fresh docs at implementation time reduces drift and avoids implementing against stale assumptions.
**Requirements:**
- Use official Google Developers documentation as the primary source.
- Fetch current docs during implementation (not only prior notes/memory).
- Record consulted URLs and access date in implementation notes/PR.
- If newly fetched docs conflict with this design, update proposal/design before implementation proceeds.

### 10. FC Centrum styleguide as normative UI source
**Choice:** Treat `docs/fccentrum-styleguide.md` as the normative visual and interaction spec for `viz` UI surfaces in this change (excluding marker data-encoding colors).
**Rationale:** The styleguide was created specifically for spots-map UI and includes implementation-ready tokens, typography hierarchy, component/state behavior, responsive layouts, and accessibility baseline/checklist. Using it as the source of truth prevents ad-hoc styling and keeps map-adjacent UI aligned with FC Centrum brand patterns.
**Implementation notes:**
- Define styleguide tokens as CSS custom properties in `viz` global styles.
- Build author filter panel and any spot list/card surfaces using the documented card/button/list/header patterns and interaction states.
- Follow the documented mobile/tablet/desktop layout behavior for map + panel composition.
- If implementation must diverge, record deviations in change/PR notes using the styleguide template.

## Risks / Trade-offs

- **[Advanced Markers require Map ID]** → A cloud-based map style must be created in Google Cloud Console before the app works. Mitigation: document the setup steps clearly.
- **[Color palette wrapping]** → With ~40-50 authors and 12-16 palette colors, multiple authors share colors. Mitigation: acceptable because the filter UI disambiguates; spots from different authors rarely overlap spatially.
- **[Stale data]** → The static JSON is only as fresh as the latest exported snapshot. Mitigation: refresh/regenerate JSON as part of data update workflow and re-deploy.
- **[API key exposure]** → The Maps API key is embedded in client-side JavaScript. Mitigation: restrict the key to the Maps JavaScript API and limit by HTTP referrer in Google Cloud Console.
