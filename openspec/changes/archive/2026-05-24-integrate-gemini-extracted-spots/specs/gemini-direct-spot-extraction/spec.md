## MODIFIED Requirements

### Requirement: Gemini-direct extraction requests structured spot output
The Gemini-direct extraction flow SHALL request structured model output containing the presenter, video identity, and timestamped spot candidates needed for persistence and comparison.

#### Scenario: Gemini-direct prompt is sent
- **WHEN** the Gemini-direct extraction flow sends a request to Gemini
- **THEN** the request SHALL require a native structured JSON response containing an optional presenter name and a list of spot candidates
- **AND** it SHALL provide the associated YouTube URL as a video file input when available
- **AND** it SHALL enable Gemini URL context for URL-based source retrieval
- **AND** each spot candidate SHALL include a non-empty place name, a YouTube timestamp in seconds when available, and a human-readable evidence field supporting the candidate
- **AND** each spot candidate MAY include a confidence value between 0 and 1
- **AND** the response contract SHALL include article and YouTube context fields sufficient to validate that the response belongs to the requested article/video pair
- **AND** the request SHALL target the configured Gemini model identifier

#### Scenario: Model response is malformed
- **WHEN** Gemini returns output that cannot be parsed into the expected structured contract
- **THEN** the flow SHALL preserve the raw response artifact
- **AND** it SHALL return an explicit parse/validation diagnostic for that article

### Requirement: Gemini-direct extraction does not mutate canonical spot export data
The Gemini-direct extraction flow SHALL persist accepted candidates to Gemini-specific SQLite tables and SHALL NOT insert those candidates into transcript-derived spot mention, geocode, article-link, or correction tables.

#### Scenario: Gemini-direct extraction produces candidates
- **WHEN** Gemini-direct extraction produces parsed spot candidates
- **THEN** it SHALL validate the response against the requested article/video context before persistence
- **AND** it SHALL insert or update accepted candidates in `gemini_direct_spot_mentions`
- **AND** it SHALL NOT insert those candidates into transcript-derived `spot_mentions`
- **AND** it SHALL NOT write transcript-derived `spot_google_geocodes`, `article_spots`, or `spot_corrections` rows

#### Scenario: Gemini-direct extraction is rerun for an article
- **WHEN** Gemini-direct extraction persists a candidate whose article source and place already exist in `gemini_direct_spot_mentions`
- **THEN** the system SHALL update or preserve the existing logical Gemini mention idempotently
- **AND** it SHALL avoid creating duplicate Gemini mention rows for the same article source and place

## ADDED Requirements

### Requirement: Gemini-direct extraction persists valid parsed responses
The Gemini-direct extraction stage SHALL write successful parsed Gemini responses directly to SQLite as part of the extraction run while retaining artifacts only for diagnostics and review.

#### Scenario: Parsed response passes validation
- **WHEN** a Gemini-direct response parses successfully and contains one or more valid spot candidates
- **THEN** the stage SHALL persist every valid candidate regardless of confidence score
- **AND** it SHALL store confidence when Gemini provides it
- **AND** it SHALL store the configured model identifier with the persisted Gemini mention when available
- **AND** it SHALL continue writing prompt, raw response, and parsed artifacts for observability

#### Scenario: Response contains diagnostics or invalid context
- **WHEN** the parsed Gemini-direct response reports diagnostics or does not match the requested article/video context
- **THEN** the stage SHALL NOT persist accepted Gemini spot mentions for that response
- **AND** it SHALL preserve available artifacts and diagnostics for review
- **AND** it SHALL return a stage error or skipped-article diagnostic with actionable context

### Requirement: Gemini-direct extraction validates candidate fields before persistence
The Gemini-direct extraction stage SHALL reject invalid candidate fields before writing Gemini-derived spot mention rows.

#### Scenario: Candidate has required values
- **WHEN** a candidate has a non-empty place name, an absent or non-negative YouTube timestamp, and an absent or in-range confidence value
- **THEN** the candidate SHALL be eligible for persistence to `gemini_direct_spot_mentions`
- **AND** place names and presenter names SHALL be trimmed before storage

#### Scenario: Candidate has invalid values
- **WHEN** a candidate has an empty place name, a negative YouTube timestamp, or a confidence value outside the range 0 through 1
- **THEN** the stage SHALL reject that candidate before database persistence
- **AND** the diagnostic SHALL identify the invalid field

### Requirement: Gemini-direct extraction records presenter linkage when present
The Gemini-direct extraction stage SHALL persist presenter attribution from valid Gemini responses using shared presenter storage without requiring transcript extraction to have run.

#### Scenario: Presenter name is present
- **WHEN** a valid Gemini-direct response includes a non-empty presenter name
- **THEN** the stage SHALL upsert the presenter name in shared presenter storage
- **AND** it SHALL link the presenter to the article source for downstream export and ordering

#### Scenario: Presenter name is empty
- **WHEN** a Gemini-direct response omits the presenter or provides only whitespace
- **THEN** the stage SHALL ignore the presenter value
- **AND** it SHALL still persist otherwise valid spot candidates
