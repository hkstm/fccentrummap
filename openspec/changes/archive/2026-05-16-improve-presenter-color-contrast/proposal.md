## Why

Presenter marker colors are currently deterministic, but adjacent presenters in the filter can receive visually similar or repeated colors as users enable presenters one by one. Improving the assignment strategy will make incremental filtering easier to scan while preserving stable presenter colors across page loads.

## What Changes

- Replace purely alphabetical palette assignment with a deterministic presenter color strategy that prioritizes contrast across the visible/filter order.
- Keep presenter colors stable for the same exported presenter list across sessions and reloads.
- Use a fixed, map-legible high-distinction palette and a deterministic assignment algorithm that spreads nearby filter entries across different hue families before repeating colors.
- Preserve the existing marker and filter UI contracts; no exported JSON fields or backend APIs are added.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `custom-markers`: Presenter marker color assignment is refined to maximize practical distinction for adjacent/filter-order presenters while remaining deterministic.

## Impact

- Affected frontend code: `viz/lib/color.ts` and tests for presenter color mapping.
- Affected UI behavior: marker colors and checked-presenter icon colors may change, but remain deterministic.
- No scraper, database, export JSON schema, Google Maps API integration, or backend changes expected.
