# custom-markers Specification

## Purpose
Define the custom marker system so FC Centrum map spots use a consistent Amsterdam X visual with deterministic, distinguishable presenter color encoding.
## Requirements
### Requirement: Amsterdam X marker rendering
The system SHALL render each spot marker as an Amsterdam "X" (andreaskruis) visual.

#### Scenario: Marker visual shape
- **WHEN** spots are rendered on the map
- **THEN** each marker SHALL use the custom Amsterdam X marker design instead of default Google pin styling

### Requirement: Deterministic presenter color mapping
The system SHALL assign marker colors deterministically by presenter using a fixed distinction palette.

#### Scenario: Stable color across sessions
- **WHEN** the same presenter appears across page loads
- **THEN** that presenter SHALL receive the same marker color each time

### Requirement: Accessible distinction palette
The system SHALL use a marker palette that prioritizes visual distinction and contrast.

#### Scenario: Palette constraints
- **WHEN** marker colors are defined
- **THEN** near-indistinguishable hues SHALL be avoided and color choices SHALL maintain practical map-legibility contrast

### Requirement: Custom marker implementation SHALL be based on latest official docs
Custom marker implementation steps SHALL be derived from the latest official Google Maps documentation and official code samples fetched from the internet.

#### Scenario: Official samples drive marker implementation
- **WHEN** implementing or changing Advanced Marker setup, marker content rendering, or marker interaction wiring
- **THEN** the implementer SHALL fetch current official Google docs/samples and implement using those patterns

#### Scenario: Multiple marker implementation paths are documented
- **WHEN** Google docs show multiple valid approaches for custom markers
- **THEN** the implementer SHALL select the best-supported/newest approach (or best-fit with explicit rationale), and document URLs plus access date in change/PR notes

