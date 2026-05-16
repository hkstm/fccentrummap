## ADDED Requirements

### Requirement: Collapsible presenter filter presentation
The system SHALL provide the presenter ("Spots van") filter in a collapsible container that is responsive across viewport sizes.

#### Scenario: Mobile default collapsed state
- **WHEN** the map loads on a mobile viewport
- **THEN** the presenter filter SHALL render collapsed by default

#### Scenario: Desktop default expanded state
- **WHEN** the map loads on a desktop viewport
- **THEN** the presenter filter SHALL render expanded by default

#### Scenario: Expand and collapse control
- **WHEN** the user activates the presenter filter toggle
- **THEN** the filter panel SHALL alternate between collapsed and expanded states

### Requirement: Collapsed state preserves map visibility and filter awareness
The system SHALL prevent the presenter filter from taking over the mobile viewport while still exposing current filter state.

#### Scenario: Mobile collapsed footprint
- **WHEN** the presenter filter is collapsed on mobile
- **THEN** the UI SHALL show only a compact header/toggle area and keep the map content visible

#### Scenario: Collapsed state summary
- **WHEN** the presenter filter is collapsed
- **THEN** the UI SHALL display a summary of active presenter selection state

#### Scenario: Existing filter semantics preserved
- **WHEN** the user expands the filter and changes selections
- **THEN** multi-select behavior, default all-selected behavior, and select-all/deselect-all behavior SHALL remain consistent with existing author filtering requirements
