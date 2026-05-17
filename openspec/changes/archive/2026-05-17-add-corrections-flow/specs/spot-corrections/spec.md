## ADDED Requirements

### Requirement: Maintainer correction command
The system SHALL provide an interactive maintainer CLI flow for creating or updating a spot correction by stable spot identifier. The command SHALL accept either the raw stable spot identifier or a map URL containing the URL-encoded `spot` query parameter. The command SHALL support hiding a spot from exported visualization data. The first implementation SHALL NOT require non-interactive edit flags.

#### Scenario: Correction flow starts from spot identifier
- **WHEN** a maintainer runs the correction command with a valid stable spot identifier from map share state
- **THEN** the command SHALL load the matching source spot and show the current effective values for spot name, Google place ID, and YouTube timestamp

#### Scenario: Correction flow starts from map URL
- **WHEN** a maintainer runs the correction command with a map URL containing `spot=2%3A20`
- **THEN** the command SHALL decode the query parameter and target spot identifier `2:20`

#### Scenario: Blank correction input preserves existing value
- **WHEN** the correction command prompts for updated spot name, place ID, or timestamped YouTube URL and the maintainer submits empty input for a field
- **THEN** the system SHALL keep the existing effective value for that field

#### Scenario: Unknown spot identifier fails clearly
- **WHEN** the correction command is run with a spot identifier that does not match an exportable source spot
- **THEN** the command SHALL fail with an actionable error indicating the spot was not found

#### Scenario: Hide spot
- **WHEN** a maintainer runs the correction command with the hide flag for a valid stable spot identifier
- **THEN** the system SHALL store a correction marking that spot hidden
- **AND** later JSON export SHALL omit that spot

### Requirement: Correction storage is separate from extracted data
The system SHALL store manual corrections separately from extracted spot, article, and geocode source records, including whether a spot is hidden from exported visualization data.

#### Scenario: New correction is saved
- **WHEN** a maintainer saves a correction for a spot identifier
- **THEN** the system SHALL persist the override fields in correction storage keyed by that stable spot identifier
- **AND** it SHALL NOT mutate the original extracted spot mention, article, or geocode source rows

#### Scenario: Existing correction is updated
- **WHEN** a maintainer saves another correction for the same stable spot identifier
- **THEN** the system SHALL update the existing correction row instead of creating duplicate active corrections

### Requirement: Corrected places resolve IDs and coordinates
The system SHALL resolve corrected Google place inputs to a Google place ID, latitude, and longitude before corrected spot data is exported. Corrected place input MAY be a Google place ID, Google Maps URL, or plain address/place query.

#### Scenario: Corrected place ID resolves successfully
- **WHEN** a maintainer provides a new Google place ID for a correction
- **THEN** the system SHALL resolve that place ID to latitude and longitude
- **AND** it SHALL store the resolved place ID and corrected coordinates with the correction

#### Scenario: Corrected Google Maps URL or address resolves successfully
- **WHEN** a maintainer provides a Google Maps URL or plain address for a correction
- **THEN** the system SHALL resolve it to a Google place ID, latitude, and longitude
- **AND** it SHALL store the resolved place ID and corrected coordinates with the correction

#### Scenario: Corrected place input cannot be resolved
- **WHEN** a maintainer provides a Google place input that cannot be resolved to coordinates
- **THEN** the correction save SHALL fail with an actionable error
- **AND** the previous correction state SHALL remain unchanged

### Requirement: Corrected timestamps update exported YouTube links
The system SHALL allow a correction to override the timestamp used in the exported YouTube link without changing the source article/video record. Timestamp correction input SHALL be a full timestamped YouTube URL using either `youtube.com` or `youtu.be`.

#### Scenario: Timestamped youtube.com URL is supplied
- **WHEN** a maintainer saves a timestamp correction using a full `youtube.com` URL containing a timestamp
- **THEN** the system SHALL extract and store the corrected timestamp
- **AND** the exported `youtubeLink` for that spot SHALL use the corrected timestamp on the existing source video URL

#### Scenario: Timestamped youtu.be URL is supplied
- **WHEN** a maintainer saves a timestamp correction using a full `youtu.be` URL containing a timestamp
- **THEN** the system SHALL extract and store the corrected timestamp
- **AND** the exported `youtubeLink` for that spot SHALL use the corrected timestamp on the existing source video URL

#### Scenario: YouTube URL without timestamp is rejected
- **WHEN** a maintainer provides a `youtube.com` or `youtu.be` URL without a timestamp
- **THEN** the correction save SHALL fail with an actionable error indicating that a timestamped YouTube URL is required

#### Scenario: Non-YouTube timestamp input is rejected
- **WHEN** a maintainer provides raw seconds, `MM:SS`, `HH:MM:SS`, or a non-YouTube URL as timestamp correction input
- **THEN** the correction save SHALL fail with an actionable error indicating that a full timestamped `youtube.com` or `youtu.be` URL is required

#### Scenario: Timestamp correction is omitted
- **WHEN** no timestamp correction exists for a spot
- **THEN** the exported `youtubeLink` SHALL keep the timestamp from the source export data
