## ADDED Requirements

### Requirement: Spot extraction operates on transcript content
The spot extraction stage SHALL use timestamped transcript content as the primary input for extracting authors, spot names, and addresses.

#### Scenario: Transcript-based extraction
- **WHEN** a transcript is available for an article's embedded video
- **THEN** the extractor SHALL derive structured `{author_name, spots}` output from that transcript content

### Requirement: Extracted spots include timestamps
Structured extraction output SHALL include a timestamp for each extracted spot when the spot is tied to a point in the transcript.

#### Scenario: Spot mentioned in transcript
- **WHEN** the extractor identifies a recommended spot from timestamped transcript text
- **THEN** the result SHALL include `timestamp_seconds` for that spot

### Requirement: Article-text extraction is not required in the current direction
The transcript-first extraction path SHALL not depend on preserving the earlier article-text extraction implementation.

#### Scenario: Transcript-first direction
- **WHEN** contributors work on this capability
- **THEN** they SHALL treat transcript-based extraction as the primary extraction path rather than extending the removed text-only flow
