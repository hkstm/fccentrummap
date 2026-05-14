## ADDED Requirements

### Requirement: Spots surface theme mapping
The system SHALL map styleguide tokens and component patterns to spots-focused UI surfaces, including map-adjacent cards/lists, section headers, metadata labels, and CTA/link elements.

#### Scenario: Map-adjacent list uses mapped styles
- **WHEN** rendering a spot card or list item next to the map interface
- **THEN** the UI MUST apply the documented token and component mappings instead of ad-hoc styling values

### Requirement: Content hierarchy for spots entries
The system SHALL define a consistent hierarchy for spots entry content (title, excerpt, category/meta, and actions) reflecting the editorial rhythm of `fccentrum.nl/categorie/spots/`.

#### Scenario: Spot entry hierarchy is consistently rendered
- **WHEN** multiple spots are displayed in list or card form
- **THEN** each entry MUST present content fields in the same visual hierarchy and spacing order

### Requirement: Interaction and accessibility baseline
The system SHALL specify interaction and accessibility expectations for spots UI elements, including focus visibility, contrast baseline, and pointer/keyboard parity for primary actions.

#### Scenario: Interactive states are accessible
- **WHEN** a user navigates the spots interface by keyboard
- **THEN** focus indicators and action affordances MUST remain visible and operable for each interactive element

### Requirement: Deviation handling
The system SHALL require that any deliberate deviation from the documented spots style mapping be explicitly noted with rationale in the change artifacts.

#### Scenario: Non-standard styling is traceable
- **WHEN** implementation introduces a style differing from documented mapping
- **THEN** the change record MUST include the deviation, reason, and intended follow-up decision