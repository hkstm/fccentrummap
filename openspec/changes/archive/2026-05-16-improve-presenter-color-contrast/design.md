## Context

The frontend renders custom Amsterdam X markers and the presenter filter using a presenter-to-color map from `viz/lib/color.ts`. Today the map is deterministic because presenter names are alphabetically sorted before assigning colors from a fixed palette. However, the filter now follows editorial/export order, and users may enable presenters one by one from that order. Alphabetic color assignment can place nearby filter entries on visually similar colors, which weakens the intended distinction between incrementally selected presenters.

The export already provides presenters in the desired UI order. The frontend can use that ordered list directly when building the color map; no backend or JSON schema change is needed.

## Goals / Non-Goals

**Goals:**
- Keep marker colors deterministic for unchanged exported presenter data.
- Improve visual distinction between adjacent presenters in the filter/export order.
- Preserve a fixed, map-legible palette and avoid new runtime dependencies.
- Keep marker components, Google Maps integration, and exported JSON shape unchanged.
- Cover the color assignment behavior with focused frontend tests.

**Non-Goals:**
- Guaranteeing a presenter receives the same color forever if the exported presenter list/order changes.
- Adding user-configurable colors or persisted client preferences.
- Changing marker shape, marker interaction behavior, map viewport behavior, or scraper/export models.
- Introducing a color-contrast library or dynamic perceptual color generation.

## Decisions

1. **Assign colors from presenter/filter order instead of alphabetic name order**
   - **Decision:** `buildPresenterColorMap` SHALL consume the presenter array in its given order, which is the UI filter/export order.
   - **Rationale:** The user experiences colors by walking down the filter list. Assigning in that order makes adjacent selections intentionally distinct.
   - **Alternatives considered:**
     - Keep alphabetical assignment: stable but disconnected from the current interaction order.
     - Hash presenter names directly to colors: stable across list changes, but adjacent filter entries can collide or land on similar hues.

2. **Use a high-distinction palette ordered for adjacent contrast**
   - **Decision:** Keep a fixed palette, but order it so neighboring slots alternate between distinct hue families and practical map-legible colors.
   - **Rationale:** A curated palette is predictable, dependency-free, and easier to test than generated colors.
   - **Alternatives considered:**
     - Golden-ratio HSL generation: flexible but can produce colors with poor legibility on map tiles unless constrained carefully.
     - Very large palette from arbitrary color lists: reduces repetition but may include low-contrast or brand-inconsistent colors.

3. **Accept palette reuse after palette exhaustion**
   - **Decision:** If presenters exceed the palette size, colors SHALL wrap deterministically.
   - **Rationale:** The current UI already has more presenters than the palette size; wrapping is simple and predictable. The primary improvement is maximizing distinction for the first visible/adjacent run before reuse.
   - **Alternatives considered:**
     - Generate unlimited colors: larger scope and harder to validate for contrast.
     - Fail or warn on palette exhaustion: not appropriate for normal frontend rendering.

## Risks / Trade-offs

- **[Trade-off] A presenter color can change when exported presenter order changes** → **Mitigation:** The order reflects current editorial recency/filter priority; unchanged data remains stable across reloads.
- **[Risk] Palette reuse can still create same-color presenters in large selections** → **Mitigation:** Keep the palette length reasonably larger than the default selected presenter count and order the palette for high adjacent contrast before wrapping.
- **[Risk] Color distinction is subjective and display-dependent** → **Mitigation:** Use a curated palette with visibly different hue families and add tests for deterministic, order-based assignment rather than trying to encode subjective perception in tests.
