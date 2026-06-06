## ADDED Requirements

### Requirement: Comparison report reads baseline and Gemini-direct results
The system SHALL generate a comparison report from existing transcript-based extraction data and Gemini-direct parsed artifacts without invoking Gemini or Google Places during report generation.

#### Scenario: Inputs are available
- **WHEN** the comparison report command is run with a SQLite database containing baseline extraction data and a directory containing Gemini-direct parsed artifacts
- **THEN** it SHALL load baseline presenter and spot data from SQLite
- **AND** it SHALL load Gemini-direct presenter and spot candidates from the parsed artifacts
- **AND** it SHALL NOT send Gemini model requests
- **AND** it SHALL NOT send Google Places geocoding requests

#### Scenario: Gemini-direct artifact is malformed
- **WHEN** a Gemini-direct parsed artifact cannot be read or parsed
- **THEN** the report SHALL include an article-level note describing the parse failure
- **AND** it SHALL continue reporting other readable articles when possible

### Requirement: Comparison report lists baseline and Gemini-direct spots without automatic matching
The comparison report SHALL list baseline spots and Gemini-direct spots per article for manual or AI-assisted review, without classifying spot matches or requiring new geocoding for Gemini-direct candidates.

#### Scenario: Baseline and Gemini-direct spots exist for the same article
- **WHEN** the report includes an article with both baseline spots and Gemini-direct spot candidates
- **THEN** it SHALL display the baseline spots with timestamps in a baseline section
- **AND** it SHALL display the Gemini-direct spots with timestamps and evidence or notes in a Gemini-direct section
- **AND** it SHALL NOT classify spots as matched, baseline-only, or Gemini-direct-only

#### Scenario: One output has no spots
- **WHEN** either the baseline output or Gemini-direct output has no spots for an article
- **THEN** the report SHALL still include that output section with an explicit empty placeholder

### Requirement: Comparison report is a Markdown review document
The system SHALL write a human-readable Markdown report organized by article with summary counts and separate baseline and Gemini-direct spot lists.

#### Scenario: Markdown report is generated
- **WHEN** comparison report generation completes
- **THEN** it SHALL write a Markdown document to the configured output path
- **AND** the document SHALL include report metadata, summary counts, and one section per listed article
- **AND** each article section SHALL include article URL and YouTube URL when available
- **AND** each article section SHALL include a presenter table with baseline and Gemini-direct values
- **AND** each article section SHALL include a baseline spots table
- **AND** each article section SHALL include a Gemini-direct spots table with evidence or notes when present
