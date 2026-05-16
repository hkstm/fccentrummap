## 1. Filter container and responsive state

- [x] 1.1 Identify the existing "Spots van" filter component and wrap its controls in a collapsible/disclosure container
- [x] 1.2 Implement viewport-based initial state (collapsed on mobile, expanded on desktop) with deterministic behavior on first load
- [x] 1.3 Ensure expanded/collapsed state toggles correctly from a semantic button and keeps existing filter state intact

## 2. Collapsed UI and accessibility

- [x] 2.1 Add a compact collapsed header that clearly labels "Spots van" and includes expand/collapse affordance
- [x] 2.2 Implement collapsed summary text for active presenter selection state (e.g., selected count/all selected)
- [x] 2.3 Add accessibility attributes and UX details (`aria-expanded`, `aria-controls`, focus styles, touch target sizing)

## 3. Behavior parity and validation

- [x] 3.1 Verify multi-select, default all-selected, and select-all/deselect-all behaviors remain unchanged inside expanded panel
- [x] 3.2 Validate on mobile that collapsed mode no longer takes over the viewport and map remains visible/usable
- [x] 3.3 Add or update UI interaction tests for mobile collapsed default, desktop expanded default, toggle behavior, and summary rendering
