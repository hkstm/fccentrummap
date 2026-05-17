## ADDED Requirements

### Requirement: JSON export applies spot corrections
The JSON export process SHALL apply stored spot corrections to exported spot records while preserving existing output compatibility for consumers.

#### Scenario: Name correction overrides exported spot name
- **WHEN** an exportable spot has a stored spot name correction
- **THEN** the exported spot record SHALL use the corrected `spotName`

#### Scenario: Place correction overrides exported place and coordinates
- **WHEN** an exportable spot has a stored Google place ID correction with resolved coordinates
- **THEN** the exported spot record SHALL use the corrected `placeId`, `latitude`, and `longitude`

#### Scenario: Timestamp correction overrides exported YouTube timestamp
- **WHEN** an exportable spot has a stored timestamp correction extracted from a timestamped `youtube.com` or `youtu.be` URL
- **THEN** the exported `youtubeLink` SHALL point to the same source video with the corrected timestamp

#### Scenario: Hidden correction omits spot
- **WHEN** an exportable spot has a stored hidden correction
- **THEN** the exported `spots` array SHALL omit that spot

#### Scenario: Empty correction fields preserve source values
- **WHEN** a correction exists but a supported override field is empty or null
- **THEN** the exported spot record SHALL use the original source value for that field

### Requirement: Correction export failures are actionable
The JSON export process SHALL fail instead of producing partially corrected output when required correction data is invalid.

#### Scenario: Corrected place ID lacks coordinates
- **WHEN** a spot correction contains a corrected place ID without resolved corrected coordinates
- **THEN** the export SHALL fail with an actionable error naming the affected stable spot identifier

#### Scenario: Corrected timestamp is invalid
- **WHEN** a spot correction contains an invalid timestamp value that was not extracted from a full timestamped `youtube.com` or `youtu.be` URL
- **THEN** the export SHALL fail with an actionable error naming the affected stable spot identifier
