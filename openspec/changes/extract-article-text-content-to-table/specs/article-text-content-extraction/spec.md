## MODIFIED Requirements

### Requirement: Extract main article text using Trafilatura
The extraction component SHALL parse stored raw article HTML with Trafilatura and produce cleaned article text suitable for downstream NLP/LLM use.

#### Scenario: Trafilatura returns usable content
- **WHEN** Trafilatura extraction returns cleaned content above the minimum content threshold
- **THEN** the extractor SHALL classify the extraction mode as `trafilatura`
- **AND** it SHALL return normalized text segments for persistence

#### Scenario: Trafilatura returns insufficient content
- **WHEN** Trafilatura extraction succeeds but extracted content is empty or below the minimum content threshold
- **THEN** the extractor SHALL classify the outcome as `no_match`
- **AND** it SHALL return no content rows

### Requirement: Extraction SHALL emit explicit outcome status for every processed article
The extraction flow SHALL produce an explicit outcome for each processed article to support drift monitoring and diagnostics.

#### Scenario: Extraction fails due to parser/runtime error
- **WHEN** extraction cannot be completed due to parser/runtime failure
- **THEN** the extraction outcome SHALL be recorded as `error`
- **AND** diagnostic context SHALL be recorded for troubleshooting
