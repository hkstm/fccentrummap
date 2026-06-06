## ADDED Requirements

### Requirement: Filter status category tabs
The system SHALL present the filter status area as a horizontally scrollable tab list with fixed tabs for presenter and category filters.

#### Scenario: Fixed status tab order
- **WHEN** the filter status area is rendered
- **THEN** the first status tab SHALL be `SPOTS VAN`
- **AND** the second status tab SHALL be `CATEGORIEËN`

#### Scenario: Category tab remains available without category data
- **WHEN** the frontend has zero category filter options
- **THEN** the filter status area SHALL still render the `CATEGORIEËN` tab
- **AND** the category additional status SHALL read `0 van 0 categorieën`

#### Scenario: Active status tab displays primary and additional status
- **WHEN** a status tab is active
- **THEN** the tab SHALL display its primary status label
- **AND** it SHALL display its additional status text

#### Scenario: Inactive status tab deemphasizes additional status
- **WHEN** a status tab is inactive
- **THEN** the tab SHALL display its primary status label in a visually greyed-out state
- **AND** its additional status text SHALL be visually hidden
- **AND** the transition between active and inactive states SHALL use CSS transitions

#### Scenario: Category additional status uses plural wording
- **WHEN** the category status tab displays selected and total category counts
- **THEN** the additional status SHALL use the format `X van Y categorieën`
- **AND** it SHALL use `categorieën` for all counts, including one and zero

### Requirement: Status tabs control visible filter options
The system SHALL use the active filter status tab to determine which filter options section is visible or focused in the expanded filter panel.

#### Scenario: Presenter tab shows presenter options
- **WHEN** the user activates the `SPOTS VAN` status tab while the filter panel is expanded
- **THEN** the panel SHALL show or scroll to the presenter filter options
- **AND** presenter bulk actions SHALL be available

#### Scenario: Category tab shows category options
- **WHEN** the user activates the `CATEGORIEËN` status tab while the filter panel is expanded
- **THEN** the panel SHALL show or scroll to the category filter options
- **AND** category bulk actions SHALL be available when category filtering actions are configured

#### Scenario: Status tab activation preserves filter selections
- **WHEN** the user switches between the `SPOTS VAN` and `CATEGORIEËN` status tabs
- **THEN** existing presenter selections SHALL be preserved
- **AND** existing category selections SHALL be preserved

#### Scenario: User verification covers tab interaction
- **WHEN** this change is implemented
- **THEN** verification SHALL include Playwright interaction with the status tabs
- **AND** verification SHALL confirm scrolling or swiping the status tab area behaves as a user-facing horizontal tab control

## MODIFIED Requirements

### Requirement: Collapsible presenter filter presentation
The system SHALL provide the presenter and category filters in a collapsible container that is responsive across viewport sizes and controlled by a tabbed filter status area.

#### Scenario: Mobile default collapsed state
- **WHEN** the map loads on a mobile viewport
- **THEN** the filter panel SHALL render collapsed by default

#### Scenario: Desktop default expanded state
- **WHEN** the map loads on a desktop viewport
- **THEN** the filter panel SHALL render expanded by default

#### Scenario: Expand and collapse control
- **WHEN** the user activates the filter panel toggle or otherwise taps/clicks the filter status area without selecting a different tab
- **THEN** the filter panel SHALL alternate between collapsed and expanded states

#### Scenario: Status tab selection remains operable
- **WHEN** the user activates a status tab in the filter status area
- **THEN** the active status tab SHALL change predictably
- **AND** the expand/collapse behavior SHALL NOT prevent tab selection

### Requirement: Collapsed state preserves map visibility and filter awareness
The system SHALL prevent the filter panel from taking over the mobile viewport while still exposing current presenter and category filter state through the tabbed filter status area.

#### Scenario: Mobile collapsed footprint
- **WHEN** the filter panel is collapsed on mobile
- **THEN** the UI SHALL show only a compact status/toggle area and keep the map content visible

#### Scenario: Collapsed state presenter summary
- **WHEN** the filter panel is collapsed and the `SPOTS VAN` tab is active
- **THEN** the UI SHALL display the presenter primary status
- **AND** it SHALL display the current active presenter selection summary

#### Scenario: Collapsed state category summary
- **WHEN** the filter panel is collapsed and the `CATEGORIEËN` tab is active
- **THEN** the UI SHALL display the category primary status
- **AND** it SHALL display the current active category selection summary using the format `X van Y categorieën`

#### Scenario: Existing filter semantics preserved
- **WHEN** the user expands the filter and changes selections
- **THEN** multi-select behavior, default all-selected behavior, and select-all/deselect-all behavior SHALL remain consistent with existing author filtering requirements
