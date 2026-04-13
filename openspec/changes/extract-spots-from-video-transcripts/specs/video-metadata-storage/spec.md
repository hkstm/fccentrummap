## ADDED Requirements

### Requirement: Spot records may store source video metadata
The data model SHALL support storing source video metadata for spots extracted from embedded videos.

#### Scenario: Spot extracted from video transcript
- **WHEN** a spot is derived from a transcript-backed video source
- **THEN** the stored spot data SHALL support associating that spot with a `video_url` and `timestamp_seconds`

### Requirement: Video metadata supports deep-linking
Stored video metadata SHALL be sufficient to open the source video near the moment where the spot was mentioned.

#### Scenario: Future marker deep link
- **WHEN** a client later constructs a link to the source video for a stored spot
- **THEN** the stored metadata SHALL be sufficient to build a YouTube link with a seconds-offset parameter
