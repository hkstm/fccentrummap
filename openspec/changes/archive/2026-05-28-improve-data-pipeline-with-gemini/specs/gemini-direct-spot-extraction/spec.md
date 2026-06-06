## MODIFIED Requirements

### Requirement: Gemini-direct extraction uses article and YouTube URLs only
The system SHALL provide the primary Gemini-direct extraction flow that builds model input from the stored article URL and associated YouTube URL, without requiring cleaned article text or transcription output.

#### Scenario: Article and YouTube URLs are available
- **WHEN** the Gemini-direct extraction flow selects an article with a stored article URL and associated YouTube URL
- **THEN** it SHALL compose a prompt containing those URLs as the primary source inputs
- **AND** it SHALL NOT require cleaned article text to be present
- **AND** it SHALL NOT require Murmel transcription output to be present

#### Scenario: Required URL is missing
- **WHEN** the Gemini-direct extraction flow selects an article without an article URL or without an associated YouTube URL
- **THEN** it SHALL fail or skip that article with an explicit diagnostic naming the missing input
- **AND** it SHALL NOT send a Gemini request for that article

### Requirement: Gemini-direct extraction requests structured spot output
The Gemini-direct extraction flow SHALL request structured model output containing the presenter, video identity, optional spot addresses, and timestamped spot candidates needed for persistence and downstream processing.

#### Scenario: Gemini-direct prompt is sent
- **WHEN** the Gemini-direct extraction flow sends a request to Gemini
- **THEN** the request SHALL require a native structured JSON response containing an optional presenter name and a list of spot candidates
- **AND** it SHALL instruct Gemini to use the article title and video title to determine the canonical presenter name and spelling, preferring the title form when it represents the name the person is known by
- **AND** it SHALL provide the associated YouTube URL as a video file input when available
- **AND** it SHALL enable Gemini URL context for URL-based source retrieval
- **AND** each spot candidate SHALL include a non-empty place name, a YouTube timestamp in seconds when available, and a human-readable evidence field supporting the candidate
- **AND** each spot candidate SHALL include an address only when Gemini is reasonably confident, including transition-card subtitles when they clearly represent an address
- **AND** each spot candidate MAY include a confidence value between 0 and 1
- **AND** the response contract SHALL include article and YouTube context fields sufficient to validate that the response belongs to the requested article/video pair
- **AND** the request SHALL target the configured Gemini model identifier

#### Scenario: Model response is malformed
- **WHEN** Gemini returns output that cannot be parsed into the expected structured contract
- **THEN** the flow SHALL preserve the raw response artifact
- **AND** it SHALL return an explicit parse/validation diagnostic for that article

### Requirement: Gemini-direct extraction writes reviewable artifacts
The Gemini-direct extraction flow SHALL persist prompt, raw response, and parsed output artifacts for each attempted article so extraction runs can be reviewed without repeating model calls.

#### Scenario: Gemini-direct extraction succeeds
- **WHEN** Gemini-direct extraction completes successfully for an article
- **THEN** the flow SHALL write the exact prompt text to a deterministic artifact path
- **AND** it SHALL write the raw Gemini response body to a deterministic artifact path
- **AND** it SHALL write the parsed candidate JSON to a deterministic artifact path
- **AND** the parsed artifact SHALL include the article URL, YouTube URL, model identifier, presenter value, and spot candidates with nullable address values

#### Scenario: Gemini-direct extraction fails after prompt creation
- **WHEN** a Gemini request or response parsing step fails after the prompt has been composed
- **THEN** the flow SHALL preserve every artifact that was available before the failure
- **AND** it SHALL avoid silently discarding diagnostics needed for manual review

### Requirement: Gemini-direct extraction persists valid parsed responses
The Gemini-direct extraction stage SHALL write successful parsed Gemini responses directly to SQLite as part of the extraction run while retaining artifacts only for diagnostics and review.

#### Scenario: Parsed response passes validation
- **WHEN** a Gemini-direct response parses successfully and contains one or more valid spot candidates
- **THEN** the stage SHALL persist every valid candidate regardless of confidence score
- **AND** it SHALL store the candidate address when Gemini provides a non-empty value
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
- **WHEN** a candidate has a non-empty place name, an absent or non-negative YouTube timestamp, an absent or in-range confidence value, and an absent or non-empty address value
- **THEN** the candidate SHALL be eligible for persistence to `gemini_direct_spot_mentions`
- **AND** place names, presenter names, and address values SHALL be trimmed before storage

#### Scenario: Candidate has invalid values
- **WHEN** a candidate has an empty place name, a negative YouTube timestamp, a confidence value outside the range 0 through 1, or a whitespace-only address value
- **THEN** the stage SHALL reject that candidate before database persistence
- **AND** the diagnostic SHALL identify the invalid field

### Requirement: Gemini-direct extraction does not depend on legacy transcript tables
The Gemini-direct extraction flow SHALL persist accepted candidates to Gemini-specific SQLite tables and SHALL NOT require legacy transcription-derived extraction tables to exist.

#### Scenario: Gemini-direct extraction produces candidates
- **WHEN** Gemini-direct extraction produces parsed spot candidates
- **THEN** it SHALL validate the response against the requested article/video context before persistence
- **AND** it SHALL insert or update accepted candidates in `gemini_direct_spot_mentions`
- **AND** it SHALL persist any provided address in the mention row
- **AND** it SHALL NOT require transcript-derived spot mention, geocode, article-link, correction, audio, transcription, or article-text tables

#### Scenario: Gemini-direct extraction is rerun for an article
- **WHEN** Gemini-direct extraction persists a candidate whose article source and place already exist in `gemini_direct_spot_mentions`
- **THEN** the system SHALL update or preserve the existing logical Gemini mention idempotently
- **AND** it SHALL avoid creating duplicate Gemini mention rows for the same article source and place

## ADDED Requirements

### Requirement: Gemini-derived data can be reset for full reprocessing
The system SHALL provide an operator path to remove previously generated Gemini-derived extraction, geocoding, article-link, and correction data before rerunning the pipeline from scratch.

#### Scenario: Gemini data reset is invoked
- **WHEN** an operator invokes the reset path for Gemini-derived data
- **THEN** the system SHALL delete existing Gemini spot mentions and dependent Gemini geocode, article-link, and correction rows
- **AND** it SHALL leave article source and article fetch records intact
- **AND** a subsequent extraction run SHALL regenerate Gemini mentions using the current prompt and schema
