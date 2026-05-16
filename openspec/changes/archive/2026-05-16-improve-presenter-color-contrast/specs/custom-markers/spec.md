## MODIFIED Requirements

### Requirement: Deterministic presenter color mapping
The system SHALL assign marker colors deterministically by presenter using the exported presenter/filter order and a fixed high-distinction palette. For unchanged exported presenter data, the same presenter SHALL receive the same marker color across page loads.

#### Scenario: Stable color across sessions
- **WHEN** the same exported presenter list is loaded across page loads
- **THEN** each presenter SHALL receive the same marker color each time

#### Scenario: Filter-order color assignment
- **WHEN** presenters are provided to the frontend in export/filter order
- **THEN** marker colors SHALL be assigned according to that order rather than alphabetic presenter-name order

#### Scenario: Palette reuse remains deterministic
- **WHEN** the number of presenters exceeds the fixed palette length
- **THEN** the system SHALL reuse palette colors in a deterministic sequence

### Requirement: Accessible distinction palette
The system SHALL use a marker palette that prioritizes visual distinction and contrast, with palette order chosen to maximize practical color difference for presenters that are adjacent in the filter order before colors repeat.

#### Scenario: Palette constraints
- **WHEN** marker colors are defined
- **THEN** near-indistinguishable hues SHALL be avoided and color choices SHALL maintain practical map-legibility contrast

#### Scenario: Adjacent filter entries use distinct early palette colors
- **WHEN** a user enables presenters one by one from the start of the filter list
- **THEN** the initially assigned marker colors SHALL come from visually distinct palette positions before any color is reused
