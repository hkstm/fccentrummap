## ADDED Requirements

### Requirement: Author filter panel displays all authors
The application SHALL display a filter panel listing all authors from the dataset, each with a checkbox and the author's palette color.

#### Scenario: Panel lists all authors
- **WHEN** the data is loaded
- **THEN** the filter panel SHALL list every author present in the dataset, sorted alphabetically

#### Scenario: Author entry shows color
- **WHEN** an author is listed in the filter panel
- **THEN** a color indicator (swatch or icon) matching the author's marker color SHALL be displayed next to their name

### Requirement: All authors are selected by default
On initial page load, all author checkboxes SHALL be checked (all spots visible).

#### Scenario: Initial state
- **WHEN** the page first loads
- **THEN** all author checkboxes SHALL be checked and all spots SHALL be visible on the map

### Requirement: Toggling an author shows or hides their spots
Unchecking an author's checkbox SHALL hide all markers for that author's spots. Re-checking SHALL show them again.

#### Scenario: Uncheck an author
- **WHEN** the user unchecks an author's checkbox
- **THEN** all markers for spots belonging to that author SHALL be removed from the map

#### Scenario: Re-check an author
- **WHEN** the user re-checks a previously unchecked author's checkbox
- **THEN** all markers for that author's spots SHALL reappear on the map

### Requirement: Select all and deselect all controls
The filter panel SHALL include a toggle to select all or deselect all authors at once.

#### Scenario: Deselect all
- **WHEN** the user clicks "Deselect all"
- **THEN** all author checkboxes SHALL be unchecked and all markers SHALL be hidden

#### Scenario: Select all
- **WHEN** the user clicks "Select all"
- **THEN** all author checkboxes SHALL be checked and all markers SHALL be visible

### Requirement: Filter panel is collapsible
The filter panel SHALL be collapsible so it does not permanently obstruct the map view.

#### Scenario: Collapse the panel
- **WHEN** the user collapses the filter panel
- **THEN** the panel SHALL minimize and the map SHALL occupy the freed space

#### Scenario: Expand the panel
- **WHEN** the user expands a collapsed filter panel
- **THEN** the full author list with checkboxes SHALL be visible again
