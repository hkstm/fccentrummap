# map-view Specification

## Purpose
Define expected map-view behavior for FC Centrum so implementers consistently render and initialize the Amsterdam map, viewport, and core marker interactions.
## Requirements
### Requirement: Interactive Amsterdam map view
The system SHALL render an interactive Google Map using the Maps JavaScript API and Advanced Marker support.

#### Scenario: Initial map load
- **WHEN** the map page is opened with valid demo key configuration
- **THEN** the map SHALL render successfully with Advanced Marker capability enabled

### Requirement: Default Amsterdam viewport
The system SHALL initialize the map viewport to the Amsterdam geocoding bounds.

#### Scenario: Default bounds applied
- **WHEN** the map first loads
- **THEN** the initial viewport SHALL fit bounds low `52.274525, 4.711585` and high `52.461764, 5.073559`

### Requirement: Marker click-through to source video
The system SHALL open the spot `youtubeLink` when a marker is activated.

#### Scenario: Open timestamped link
- **WHEN** a user clicks or keyboard-activates a spot marker
- **THEN** the browser SHALL open that spot's `youtubeLink` in a new tab, preserving any timestamp in the URL

### Requirement: Google Maps implementation SHALL be doc-driven
For every Google Maps API implementation step in this change, the implementation SHALL be based on the latest official Google documentation and code samples fetched from the internet.

#### Scenario: Official docs consulted for each Maps step
- **WHEN** implementing or modifying any Maps API behavior
- **THEN** the implementer SHALL fetch current official docs, choose an implementation pattern from official samples, and record the consulted URLs with access date

#### Scenario: Multiple implementation options exist
- **WHEN** official docs present multiple supported implementations
- **THEN** the implementer SHALL select the best-supported/newest approach (or best fit with clear rationale) and document that selection in change/PR notes

