## ADDED Requirements

### Requirement: Static JSON input contract
The system SHALL load spot data from a static JSON file at `/data/spots.json` using the documented contract.

#### Scenario: Valid data load
- **WHEN** `/data/spots.json` is available and valid
- **THEN** the frontend SHALL parse `spots` and `presenters` and render map/filter state from that data

#### Scenario: Contract schema definition
- **WHEN** validating `/data/spots.json`
- **THEN** the top-level payload SHALL include `spots` (array) and `presenters` (array)
- **AND** each `spots[*]` entry SHALL include required fields: `placeId` (string), `spotName` (string), `presenterName` (string), `latitude` (number), `longitude` (number), `youtubeLink` (string URL)
- **AND** each `spots[*]` entry MAY include optional fields: `articleUrl` (string URL)
- **AND** each `presenters[*]` entry SHALL include required field: `presenterName` (string)
- **AND** each `presenters[*]` entry MAY include optional fields: `id` (string), `avatarUrl` (string URL), `bio` (string)
- **AND** `spots[*].presenterName` SHALL match at least one `presenters[*].presenterName` value

### Requirement: Multi-presenter contract compatibility
The system SHALL support multiple presenters in input data.

#### Scenario: Multiple presenters present
- **WHEN** `presenters` contains more than one entry
- **THEN** filter options and presenter color mapping SHALL include all presenters without single-presenter assumptions

### Requirement: Missing or invalid data handling
The system SHALL surface clear failure states when the static JSON is missing or invalid.

#### Scenario: JSON unavailable
- **WHEN** `/data/spots.json` cannot be fetched
- **THEN** the UI SHALL display a clear error state instead of silently failing

#### Scenario: Malformed JSON
- **WHEN** `/data/spots.json` is fetched but cannot be parsed as valid JSON
- **THEN** the UI SHALL display a clear error state

#### Scenario: Non-array contract payload
- **WHEN** `/data/spots.json` contains non-array values for `spots` or `presenters`
- **THEN** the payload SHALL be treated as invalid contract data
- **AND** the UI SHALL display a clear error state instead of rendering partial marker interactions

#### Scenario: Missing required fields
- **WHEN** any `spots[*]` entry is missing required fields (`placeId`, `spotName`, `presenterName`, `latitude`, `longitude`, or `youtubeLink`) or any `presenters[*]` entry is missing `presenterName`
- **THEN** the payload SHALL be treated as invalid contract data
- **AND** the UI SHALL display a clear error state instead of rendering partial marker interactions

#### Scenario: Wrong required field types
- **WHEN** required fields are present but use incorrect types (for example, `latitude`/`longitude` not numeric or `youtubeLink` not a string URL)
- **THEN** the payload SHALL be treated as invalid contract data
- **AND** the UI SHALL display a clear error state instead of rendering partial marker interactions

#### Scenario: Missing or invalid required source link
- **WHEN** any `spots[*].youtubeLink` is missing, empty, or not a valid URL
- **THEN** the payload SHALL be treated as invalid contract data
- **AND** the UI SHALL display a clear error state instead of rendering partial marker interactions
