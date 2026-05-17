## MODIFIED Requirements

### Requirement: Spot records include required frontend fields
Each entry in `spots` MUST include `spotId`, `placeId`, `spotName`, `presenterName`, and `youtubeLink`. `spotId` SHALL be a stable identifier suitable for spot-specific links and maintainer correction workflows.

#### Scenario: Spot entry contains required fields
- **WHEN** a spot exists in the source dataset
- **THEN** the exported spot record MUST include `spotId`, `placeId`, `spotName`, `presenterName`, and `youtubeLink`

#### Scenario: Spot identifier remains stable across corrections
- **WHEN** a spot has corrected `spotName`, `placeId`, coordinates, or timestamp values
- **THEN** the exported `spotId` SHALL remain the same stable identifier for that source spot
