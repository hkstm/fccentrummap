## 1. Frontend bootstrap and configuration

- [x] 1.1 Scaffold `viz/` Next.js TypeScript app structure for static export
- [x] 1.2 Configure Google Maps PoC env wiring to use `DEMO_GOOGLE_MAPS_API_KEY` and `DEMO_MAP_ID`
- [x] 1.3 Add global style token setup aligned to `docs/fccentrum-styleguide.md`

## 2. Static data contract integration

- [x] 2.1 Define TypeScript types for `spots` and `presenters` from `/data/spots.json`
- [x] 2.2 Implement data loader for `/data/spots.json` with explicit missing/invalid JSON error states
- [x] 2.3 Verify multi-presenter datasets load without single-presenter assumptions

## 3. Google map rendering

- [x] 3.1 Implement map initialization with `@vis.gl/react-google-maps`
- [x] 3.2 Set initial viewport to Amsterdam bounds (low `52.274525,4.711585`, high `52.461764,5.073559`)
- [x] 3.3 Render one advanced marker per spot using `placeId`-backed coordinates from data

## 4. Custom marker visuals and interactions

- [x] 4.1 Implement Amsterdam X (andreaskruis) SVG marker component for Advanced Markers
- [x] 4.2 Implement deterministic presenter color mapping from a fixed distinction palette
- [x] 4.3 Implement marker activation behavior to open `youtubeLink` in a new tab with timestamp preserved

## 5. Presenter filter UI

- [x] 5.1 Implement collapsible presenter filter panel with all presenters selected by default
- [x] 5.2 Implement per-presenter toggle behavior to show/hide corresponding markers
- [x] 5.3 Implement select-all and deselect-all controls

## 6. Verification and implementation evidence

- [x] 6.1 Validate end-to-end flow: static JSON load -> map render -> filters -> marker link-out
- [x] 6.2 Validate key failure states: missing demo key/map id config and missing/invalid data file
- [x] 6.3 For every Google Maps API implementation step, fetch the latest official Google docs + code samples from the internet and base implementation on them
- [x] 6.4 When multiple official implementations exist, choose the best-supported/newest (or best-fit with rationale) and document the decision in change/PR notes
- [x] 6.5 Record consulted Google doc URLs and access dates in change/PR notes
- [x] 6.6 Run accessibility checks for focus visibility and keyboard activation parity on marker/filter interactions
