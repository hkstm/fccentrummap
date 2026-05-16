## Context

The existing presenter ("Spots van") filter is functionally correct but on mobile it can take over most or all of the viewport, reducing map usability and interrupting browsing. The proposal requires a responsive, collapsible filter presentation that prevents this takeover while preserving existing `author-filter` behavior (multi-select, default all selected, select-all/deselect-all). This is a frontend UX and interaction change with no backend or data-model impact.

## Goals / Non-Goals

**Goals:**
- Introduce a collapsible filter container for "Spots van" that is mobile-first and still usable on desktop.
- Keep current filtering semantics unchanged.
- Provide a compact collapsed summary so users can understand active filtering at a glance.
- Ensure the interaction is keyboard-accessible and screen-reader friendly.

**Non-Goals:**
- Changing filter logic, presenter data source, or map data contracts.
- Introducing new backend APIs or storage.
- Redesigning unrelated map controls.

## Decisions

1. **Use a responsive collapsible container around the existing presenter filter controls**
   - **Decision:** Wrap current filter controls in a disclosure/accordion-style UI.
   - **Rationale:** Reuses proven filtering logic while preventing mobile viewport takeover.
   - **Alternatives considered:**
     - Full-screen modal/sheet: good space management, but adds extra interaction cost.
     - Permanently reduced control density: still risks poor readability and discoverability.

2. **Default state by viewport: collapsed on mobile, expanded on desktop**
   - **Decision:** On small viewports, initialize collapsed; on desktop, initialize expanded.
   - **Rationale:** Maximizes map area on mobile while preserving quick visibility on larger screens.
   - **Alternatives considered:**
     - Always collapsed: reduces discoverability on desktop.
     - Always expanded: fails mobile usability objective.

3. **Collapsed summary text shows active selection count and quick context**
   - **Decision:** Header displays a concise summary (e.g., `Spots van (3/7)` or `All presenters`).
   - **Rationale:** Users can confirm filter state without opening the panel.
   - **Alternatives considered:**
     - No summary: ambiguous state when collapsed.
     - Long list of selected names: too verbose for narrow screens.

4. **Preserve existing bulk controls and selection behavior inside expanded panel**
   - **Decision:** Keep select-all/deselect-all and per-presenter toggles unchanged.
   - **Rationale:** Maintains behavioral continuity and aligns with existing `author-filter` capability.
   - **Alternatives considered:**
     - Remove bulk controls on mobile: simpler UI but reduced usability for larger presenter lists.

5. **Accessibility-first disclosure behavior**
   - **Decision:** Use semantic button for toggle with `aria-expanded`, `aria-controls`, visible focus styles, and adequate touch targets.
   - **Rationale:** Ensures the compact UX remains operable across devices and assistive technologies.
   - **Alternatives considered:**
     - Custom non-semantic clickable container: faster to style but harms accessibility.

## Risks / Trade-offs

- **[Risk] Collapsed state may hide important controls from first-time users** → **Mitigation:** Use explicit "Spots van" label, count summary, and clear chevron/toggle affordance.
- **[Risk] Responsive breakpoint behavior may feel inconsistent during resize/orientation changes** → **Mitigation:** Define deterministic initialization and state-retention rules for viewport transitions.
- **[Risk] More UI state complexity can introduce regressions in filtering interactions** → **Mitigation:** Add interaction tests for collapsed/expanded modes and verify parity with existing selection behavior.
- **[Trade-off] Slightly more component complexity for significantly better mobile map visibility** → **Mitigation:** Isolate collapse state in a wrapper component and keep filter logic untouched.
