## Why

We need a consistent visual styleguide derived from https://fccentrum.nl/ so the upcoming spots-map demo feels native to the client’s existing brand and content style. Defining this now reduces rework and de-risks implementing `render-spots-map` UI decisions against unclear design expectations.

## What Changes

- Create a reusable styleguide for typography, color, spacing, layout patterns, and UI components based on the visual language of fccentrum.nl and the `/categorie/spots/` listing style.
- Define concrete styling rules and tokens that can be applied when implementing the map and spot-related interface patterns.
- Document page-level patterns (cards/list items, section headers, CTA treatments, responsive behavior) needed to align the map experience with the existing site.
- Establish acceptance criteria for “on-brand” styling consistency in demo deliverables.

## Capabilities

### New Capabilities
- `fccentrum-styleguide`: Define brand-aligned design tokens and component-level style rules derived from fccentrum.nl.
- `spots-ui-theme-application`: Specify how the styleguide is applied to spots-focused UI surfaces (including map-adjacent list/card patterns).

### Modified Capabilities
- None.

## Deliverable

- `docs/fccentrum-styleguide.md` (token baseline, component/state rules, responsive guidance, spots mapping, accessibility/checklist, deviation logging, and `render-spots-map` handoff)

## Impact

- Affects UI design/implementation decisions for the `render-spots-map` change and related frontend presentation layers.
- Introduces new OpenSpec capability specs under this change.
- No backend API contract changes expected.
- May require frontend styling assets/configuration updates (CSS variables, component styles, design documentation).