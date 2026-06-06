## MODIFIED Requirements

### Requirement: Static JSON input contract
The system SHALL load spot data from a static JSON file at `/data/spots.json` using the documented contract.

#### Scenario: Valid data load
- **WHEN** `/data/spots.json` is available and valid
- **THEN** the frontend SHALL parse `spots`
- **AND** it SHALL derive presenter filter values from presenter names contained in the `spots` array
- **AND** it SHALL render map/filter state from that data

### Requirement: Multi-presenter contract compatibility
The system SHALL support multiple presenters in input data by deriving unique presenter names from exported spots.

#### Scenario: Multiple presenters present
- **WHEN** spots reference more than one presenter name
- **THEN** filter options and presenter color mapping SHALL include all unique presenter names without single-presenter assumptions
- **AND** duplicate presenter names across spots SHALL produce one filter option per presenter

## ADDED Requirements

### Requirement: Frontend does not require top-level presenters array
The frontend SHALL NOT require a top-level `presenters` array in `/data/spots.json`.

#### Scenario: Export omits presenters array
- **WHEN** `/data/spots.json` contains a valid `spots` array and no top-level `presenters` array
- **THEN** the frontend SHALL derive the presenter list from spots
- **AND** it SHALL render presenter filters without a data-load failure
