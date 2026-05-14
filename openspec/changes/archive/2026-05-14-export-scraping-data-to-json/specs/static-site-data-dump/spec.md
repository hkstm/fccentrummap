## ADDED Requirements

### Requirement: Static data dump document for site consumption
The system SHALL generate a single JSON document containing top-level `spots` and `presenters` collections for static site consumption.

#### Scenario: Export produces top-level collections
- **WHEN** the user runs the export command against a valid SQLite dataset
- **THEN** the output file MUST contain a JSON object with `spots` and `presenters` keys

### Requirement: Spot records include required frontend fields
Each entry in `spots` MUST include `placeId`, `spotName`, `presenterName`, and `youtubeLink`.

#### Scenario: Spot entry contains required fields
- **WHEN** a spot exists in the source dataset
- **THEN** the exported spot record MUST include `placeId`, `spotName`, `presenterName`, and `youtubeLink`

### Requirement: Presenter list is exported from database values as-is
The `presenters` collection SHALL contain presenter names taken directly from stored database values without normalization or canonicalization in v1.

#### Scenario: Presenter value preserves stored formatting
- **WHEN** a presenter name is stored in the database
- **THEN** the corresponding exported `presenterName` MUST match the stored value exactly
