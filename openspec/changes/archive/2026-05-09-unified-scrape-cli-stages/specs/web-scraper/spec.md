## MODIFIED Requirements

### Requirement: Fetch article HTML and store it as raw input
The scraper SHALL fetch article HTML, store it as raw input, detect embedded YouTube videos for downstream audio acquisition, and run local article-text extraction as part of the fetch stage.

#### Scenario: Successful fetch with embedded video
- **WHEN** an article page responds successfully and contains an embedded YouTube video
- **THEN** the scraper SHALL store the raw HTML
- **AND** it SHALL persist a nullable string field `video_id` on the raw article record
- **AND** `video_id` SHALL contain the normalized canonical YouTube ID (11-character ID only, no URL prefix)
- **AND** it SHALL persist article text extraction outcome/content for that article in the same stage flow

#### Scenario: Successful fetch without embedded video
- **WHEN** an article page responds successfully and does not contain an embedded YouTube video
- **THEN** the scraper SHALL still store the raw HTML without failing the fetch phase
- **AND** it SHALL still attempt and persist article text extraction outcome/content

#### Scenario: Respectful request pacing
- **WHEN** the scraper makes requests to `fccentrum.nl`
- **THEN** it SHALL apply a delay between requests instead of issuing them as a tight burst
