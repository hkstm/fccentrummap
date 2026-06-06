## Context

`PresenterFilterPanel.tsx` currently combines two responsibilities: expanded filter options for presenters/categories and a bottom collapsed/expanded status control that only communicates presenter selection (`Spots van` plus presenter summary). The requested change makes the bottom Filter Status area into horizontally swipeable/scrollable tabs for presenter and category filter summaries, and each active status tab controls which corresponding filter-options section is shown/scrolled into view.

The frontend already uses shadcn-style local UI primitives and Tailwind classes. The implementation should prefer local composition over introducing a new package dependency, while using the shadcn scrollable tabs example as the interaction/style reference.

## Goals / Non-Goals

**Goals:**
- Replace the single Filter Status button content with two scrollable/swipeable status tabs in fixed order: `SPOTS VAN`, then `CATEGORIEËN`.
- Always show both tabs, including when there are `0` categories.
- Make the active status tab determine the visible/relevant Filter Options section: presenter tab shows presenter options; category tab shows category options.
- Preserve the existing expand/collapse behavior of the filter panel, while allowing the whole Filter Status area to remain tappable/clickable for expansion where this does not conflict with tab selection.
- Show both primary and additional status on the active tab.
- Show greyed-out primary status on inactive tabs, with the inactive additional status visually hidden rather than removed from layout where possible.
- Use CSS transitions for inactive greying and additional-status reveal/hide.
- Use always-plural category additional status text: `X van Y categorieën`.
- Verify the behavior with Playwright by interacting with the filter status tabs and scrolling/swiping them as a user.

**Non-Goals:**
- Reworking the checkbox behavior or changing filter semantics beyond splitting the visible options by active tab.
- Changing the static data contract for presenters/categories.
- Adding new filtering dimensions beyond presenters and categories.
- Changing map marker filtering semantics.

## Decisions

1. **Use the Filter Status tabs to control the visible Filter Options section.**
   - Rationale: The intended interaction is that selecting/swiping to a status tab also focuses the matching options: `SPOTS VAN` shows presenter checkboxes and presenter bulk actions; `CATEGORIEËN` shows category checkboxes and category bulk actions.
   - Implementation approach: maintain one `activeFilterCategory` state shared by the status tabs and options area. When the panel is expanded, render only the active section or scroll the active section into view so users see the relevant options for the selected tab.
   - Alternative considered: Keep all filter options visible and change only the status summary. Rejected because it does not match the desired tabbed filtering interaction.

2. **Model filter categories as local data derived from current counts and option renderers.**
   - Rationale: Presenter and category tabs share the same rendering shape: primary label, additional summary, active state, and associated options section. A small local array keeps the JSX consistent and makes future filter categories easier to add.
   - Alternative considered: Hard-code two separate blocks. Rejected because it duplicates active/inactive transition logic.

3. **Use native horizontal overflow and button tabs instead of adding a new Tabs dependency.**
   - Rationale: The project already has local shadcn primitives but no Tabs component. The shadcn scrollable-tabs pattern can be followed with semantic buttons, `overflow-x-auto`, `snap-x`, and Tailwind transition classes without adding Radix Tabs solely for status display.
   - Alternative considered: Add shadcn/Radix Tabs. Rejected unless implementation finds an existing local Tabs primitive or a stronger accessibility need, because this status area is not switching content panels; it is selecting the active summary tab while the same filter options remain available above.

4. **Maintain one expand/collapse affordance for the panel.**
   - Rationale: The current CardHeader button toggles panel visibility. With nested tab buttons, the implementation should avoid invalid nested interactive elements. The tab strip and the chevron toggle should be sibling controls inside the header/status area.
   - Alternative considered: Make each tab also toggle expansion. Rejected because it conflates tab selection with panel expansion and may frustrate users trying to switch status tabs.

5. **Use accessible tab-like semantics where practical.**
   - Rationale: The status controls represent selectable categories. The tab row should expose labels and selected state with `role="tablist"`, `role="tab"`, and `aria-selected`, while preserving keyboard focus styles.
   - Alternative considered: Plain buttons with visual-only active state. Rejected because assistive technology should understand which status tab is active.

## Risks / Trade-offs

- **Nested interactive control regression** → Keep tab buttons separate from the expand/collapse button; do not place buttons inside another button.
- **Category summary edge cases when categories are absent** → Always render the category tab and derive totals from `categories.length`, showing `0 van 0 categorieën` when no category data exists.
- **Scrollable tabs may be hard to discover with only two tabs** → Use visible overflow-safe spacing and verify horizontal scrolling/swiping with Playwright; two tabs still establishes the requested pattern for future categories.
- **Animation layout shift** → Prefer visually hiding inactive additional status with opacity/visibility transitions over fully removing it from layout, to reduce glitches.
- **Chevron/tabs interaction ambiguity** → The whole Filter Status area may remain tappable/clickable for expansion, but active tab selection must still work predictably and must not be blocked by expand/collapse behavior.
- **Options scroll/visibility mismatch** → Keep status tab state and options visibility driven by the same `activeFilterCategory`; verify with Playwright that selecting/swiping to `CATEGORIEËN` shows category options and returning to `SPOTS VAN` shows presenter options.
