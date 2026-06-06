## MODIFIED Requirements

### Requirement: Static JSON input contract
The system SHALL load spot data from a static JSON file at `/data/spots.json` using the documented contract. The document SHALL contain exported spots with denormalized `presenterName` and `categoryName` fields and SHALL provide top-level `presenters` and `categories` arrays as ordered filter metadata.

#### Scenario: Valid data load
- **WHEN** `/data/spots.json` is available and valid
- **THEN** the frontend SHALL parse `spots`
- **AND** it SHALL parse top-level `presenters` entries containing `presenterName`
- **AND** it SHALL parse top-level `categories` entries containing `categoryName`
- **AND** it SHALL render map/filter state from that data

#### Scenario: Spot includes category name
- **WHEN** `/data/spots.json` contains a spot record
- **THEN** the spot record MUST include `categoryName`
- **AND** the frontend SHALL use `categoryName` as the public category value
- **AND** the frontend SHALL NOT require the raw primary type display name field

### Requirement: Multi-presenter contract compatibility
The system SHALL support multiple presenters in input data by consuming ordered presenter metadata and maintaining denormalized presenter names on exported spots.

#### Scenario: Multiple presenters present
- **WHEN** spots reference more than one presenter name
- **THEN** filter options and presenter color mapping SHALL include all top-level presenter names without single-presenter assumptions
- **AND** duplicate presenter names across spots SHALL correspond to one top-level presenter option per presenter
- **AND** presenter filters SHALL preserve the order provided by the top-level `presenters` array

### Requirement: Frontend does not require top-level presenters array
The frontend SHALL consume the top-level `presenters` array in `/data/spots.json` as the presenter filter metadata contract.

#### Scenario: Export includes presenters array
- **WHEN** `/data/spots.json` contains a valid `spots` array and a top-level `presenters` array
- **THEN** the frontend SHALL render presenter filters from the top-level `presenters` array
- **AND** spot records SHALL retain `presenterName` for denormalized marker, popup, and sharing behavior

#### Scenario: Export omits presenters array for empty data
- **WHEN** `/data/spots.json` contains an empty `spots` array and no top-level `presenters` array
- **THEN** the frontend SHALL treat presenter filter options as empty
- **AND** it SHALL NOT fail data loading solely because the presenter metadata array is absent for empty data

## ADDED Requirements

### Requirement: Category filter metadata contract
The system SHALL support top-level category metadata in `/data/spots.json` for category filter ordering.

#### Scenario: Categories present
- **WHEN** `/data/spots.json` contains exported spots with `categoryName`
- **THEN** the document SHALL include a top-level `categories` array
- **AND** each category entry SHALL contain `categoryName`
- **AND** category filters SHALL preserve the order provided by the top-level `categories` array

#### Scenario: Empty categories for empty data
- **WHEN** `/data/spots.json` contains an empty `spots` array
- **THEN** the frontend SHALL treat category filter options as empty when the top-level `categories` array is empty or absent

### Requirement: Static JSON category values
The system SHALL treat `categoryName` as the frontend category display and filtering value.

#### Scenario: Spot category displayed
- **WHEN** a spot record includes `categoryName`
- **THEN** the frontend SHALL use that value for category display or filtering behavior
- **AND** the frontend SHALL NOT display or filter using the raw Google Places primary type display name field
