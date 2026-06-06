## ADDED Requirements

### Requirement: Gemini-direct extraction uses article and YouTube URLs only
The system SHALL provide an experimental Gemini-direct extraction flow that builds model input from the stored article URL and associated YouTube URL, without requiring cleaned article text or transcription output.

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
The Gemini-direct extraction flow SHALL request structured model output containing the presenter and timestamped spot candidates needed for comparison.

#### Scenario: Gemini-direct prompt is sent
- **WHEN** the Gemini-direct extraction flow sends a request to Gemini
- **THEN** the request SHALL require a native structured JSON response containing an optional presenter name and a list of spot candidates
- **AND** it SHALL provide the associated YouTube URL as a video file input when available
- **AND** it SHALL enable Gemini URL context for URL-based source retrieval
- **AND** each spot candidate SHALL include a place name, a YouTube timestamp in seconds, and evidence or notes supporting the candidate
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
- **AND** the parsed artifact SHALL include the article URL, YouTube URL, model identifier, presenter value, and spot candidates

#### Scenario: Gemini-direct extraction fails after prompt creation
- **WHEN** a Gemini request or response parsing step fails after the prompt has been composed
- **THEN** the flow SHALL preserve every artifact that was available before the failure
- **AND** it SHALL avoid silently discarding diagnostics needed for manual review

### Requirement: Gemini-direct extraction does not mutate canonical spot export data
The Gemini-direct extraction flow SHALL be experimental and SHALL NOT modify canonical extraction, geocoding, correction, or export records used by the default `spots.json` export.

#### Scenario: Gemini-direct extraction produces candidates
- **WHEN** Gemini-direct extraction produces parsed spot candidates
- **THEN** it SHALL NOT insert those candidates into the canonical transcript-based spot mention tables
- **AND** it SHALL NOT change the default input data used by `geocode-spots` or `export-data`
