# render-spots-map implementation notes

## Google Maps docs consulted

Access date: 2026-05-14

- https://developers.google.com/maps/documentation/javascript/advanced-markers/html-markers
- https://developers.google.com/maps/documentation/javascript/examples/advanced-markers-html-simple
- https://developers.google.com/maps/documentation/javascript/examples/advanced-markers-basic-style
- https://visgl.github.io/react-google-maps/docs/api-reference/components/advanced-marker
- https://visgl.github.io/react-google-maps/examples/advanced-marker
- https://developers.google.com/maps/documentation/javascript/geocoding?csw=1
- https://developers.google.com/maps/documentation/javascript/examples/geocoding-place-id
- https://developers.google.com/maps/documentation/javascript/reference/geocoder

## Implementation choices

- Chosen approach: `@vis.gl/react-google-maps` with `<APIProvider>`, `<Map>`, and `<AdvancedMarker>` for React-first implementation and current official support.
- Custom marker path: Advanced Marker with custom HTML/SVG content (Amsterdam X), matching the official Advanced Marker HTML guidance.
- Coordinate resolution path (final): consume exported `latitude`/`longitude` from `/data/spots.json` directly and avoid runtime geocoding in the frontend.

## Accessibility verification (Playwright CLI)

Access date: 2026-05-14

- Keyboard activation parity verified for presenter controls:
  - Focused `Deselect all` and activated with `Enter` → checked presenters `0`, marker buttons `0`.
  - Focused `Select all` and activated with `Enter` → checked presenters `1`, marker buttons `3`.
- Keyboard activation parity verified for markers:
  - Focused first marker button and pressed `Enter` while stubbing `window.open` → captured expected timestamped URL.
- Focus visibility verified:
  - `Select all` button: computed focus outline `solid|2px`
  - presenter checkbox: computed focus outline `solid|2px`
  - marker button: explicit `.markerButton:focus/:focus-visible` outline added and verified as `solid|2px`.
