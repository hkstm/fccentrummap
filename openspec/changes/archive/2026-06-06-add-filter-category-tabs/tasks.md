## 1. Inspect Current UI and Test Surface

- [x] 1.1 Review `viz/components/PresenterFilterPanel.tsx` and identify current presenter/category option rendering, status summary rendering, and expand/collapse event handling.
- [x] 1.2 Identify available frontend test commands and Playwright CLI usage for running the local visualization and interacting with the filter panel.

## 2. Implement Tabbed Filter State

- [x] 2.1 Add local active filter category state with fixed tab order: `SPOTS VAN` first and `CATEGORIEËN` second.
- [x] 2.2 Derive presenter and category selected/total counts for status tab data, including `0 van 0 categorieën` when no categories exist.
- [x] 2.3 Ensure category additional status always uses the plural format `X van Y categorieën`.

## 3. Split Visible Filter Options by Active Tab

- [x] 3.1 Refactor the expanded Filter Options area so the presenter tab shows presenter bulk actions and presenter checkbox options.
- [x] 3.2 Refactor the expanded Filter Options area so the category tab shows category bulk actions and category checkbox options.
- [x] 3.3 Preserve existing presenter and category selection behavior when switching between tabs.
- [x] 3.4 Ensure switching active status tabs shows or scrolls to the corresponding options section without changing selected filters.

## 4. Implement Scrollable Filter Status Tabs

- [x] 4.1 Replace the single bottom status label with a horizontally scrollable/swipeable tab list using native overflow and shadcn-style/Tailwind composition.
- [x] 4.2 Add accessible tab semantics with `role="tablist"`, `role="tab"`, and selected-state attributes.
- [x] 4.3 Render active tabs with primary and additional status visible.
- [x] 4.4 Render inactive tabs with greyed-out primary status and visually hidden additional status.
- [x] 4.5 Add CSS transitions for active/inactive greying and additional-status visibility changes.
- [x] 4.6 Preserve expand/collapse behavior while ensuring status tab selection remains predictable and operable.

## 5. Validate Behavior

- [x] 5.1 Run frontend lint/type/test checks available in `viz` and fix any regressions.
- [x] 5.2 Use Playwright CLI against the local visualization to click/tap `SPOTS VAN` and verify presenter options are shown.
- [x] 5.3 Use Playwright CLI against the local visualization to click/tap `CATEGORIEËN` and verify category options are shown.
- [x] 5.4 Use Playwright CLI to horizontally scroll or swipe the filter status tab area and verify it behaves as a user-facing scrollable tab control.
- [x] 5.5 Use Playwright CLI or browser inspection to verify collapsed and expanded states keep the map visible on mobile-sized viewport.
