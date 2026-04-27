## MODIFIED Requirements

### Requirement: Extraction input uses sentence-level transcript units with timestamps
The extraction flow SHALL build model input from two sources: cleaned article text and sentence-level transcript units with sentence start timestamps. The cleaned article SHALL be the primary source for identifying candidate places, and transcript sentence units SHALL be the source of temporal anchoring.

#### Scenario: Cleaned article and sentence transcript are available
- **WHEN** the selected article has cleaned article text and transcript sentence-level units with `text` and `start` fields
- **THEN** the command SHALL compose extraction input with both sources
- **AND** the model instruction SHALL treat cleaned article text as primary for place identification
- **AND** each extracted place SHALL be traceable to a transcript sentence start timestamp

#### Scenario: Sentence-level transcript missing
- **WHEN** sentence-level transcript units are missing and only full-text blobs or word-level token timings are present
- **THEN** the command SHALL fail with a clear validation error
- **AND** it SHALL not send a model request

### Requirement: Prompt requires Dutch extraction focused on Amsterdam region
The extraction prompt SHALL be written in Dutch and SHALL instruct the model to extract places in Amsterdam or nearby surroundings, using cleaned article text as primary place source and transcript sentence units for evidence/timing.

#### Scenario: Prompt composition for extraction run
- **WHEN** the command composes the prompt from cleaned article text and transcript sentences
- **THEN** the prompt SHALL instruct the model to find "De spots van" place mentions
- **AND** it SHALL constrain candidate places to Amsterdam or nearby surroundings
- **AND** it SHALL require transcript-backed evidence/timestamp alignment for every extracted place

### Requirement: Response must be parseable into minimal extraction contract
The extraction flow SHALL validate model output against a structured contract that includes `place` and `sentenceStartTimestamp` for each extracted place.

#### Scenario: Response parses to required fields
- **WHEN** each extracted candidate includes `place` and `sentenceStartTimestamp`
- **THEN** the command SHALL treat the extraction response as valid

#### Scenario: Response cannot be parsed into required fields
- **WHEN** the model output is non-JSON or missing required fields for extraction items
- **THEN** the command SHALL fail with a parse/validation error
- **AND** it SHALL preserve the raw model response artifact for debugging
