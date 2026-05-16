## Why

The current presenter ("Spots van") filter works on desktop, but on mobile it can effectively take over the viewport and block map exploration. We should introduce a collapsible version so filtering remains accessible without a full-screen takeover on small screens.

## What Changes

- Add a collapsible "Spots van" presenter filter UI that is optimized for mobile and also usable on desktop.
- Make the filter compact by default on mobile (collapsed), with a clear affordance to expand/collapse and avoid full-screen takeover.
- Preserve existing presenter filtering behavior (multi-select, default selected, bulk controls) when used inside the collapsible container.
- Show a compact collapsed summary state so users can understand active filter state without opening the full panel.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `author-filter`: Extend requirements to support responsive, collapsible filter presentation while preserving existing presenter filtering behavior.

## Impact

- Affected frontend filter/map UI components and responsive styling.
- No backend, data model, or external dependency changes expected.
- May require updates to UI interaction tests for mobile and desktop filter behavior.
