## ADDED Requirements

### Requirement: Extraction input uses sentence-level transcript units with timestamps
The extraction flow SHALL build model input only from sentence-level transcript units that include sentence text and sentence start timestamp.

#### Scenario: Valid sentence-level transcript available
- **WHEN** the selected transcript contains sentence-level units with `text` and `start` timestamp fields
- **THEN** the command SHALL compose the extraction prompt using those sentence units
- **AND** each extracted result SHALL be traceable to a sentence start timestamp

#### Scenario: Only full-text or word-level transcript data available
- **WHEN** sentence-level transcript units are missing and only full-text blobs or word-level token timings are present
- **THEN** the command SHALL fail with a clear validation error
- **AND** it SHALL not send a model request

### Requirement: Prompt requires Dutch extraction focused on Amsterdam region
The extraction prompt SHALL be written in Dutch and SHALL instruct the model to extract places in Amsterdam or nearby surroundings.

#### Scenario: Prompt composition for extraction run
- **WHEN** the command composes the prompt from transcript sentences
- **THEN** the prompt SHALL instruct the model to find "De spots van" place mentions
- **AND** it SHALL constrain candidate places to Amsterdam or nearby surroundings

### Requirement: Candidate count is a target, not a strict acceptance gate
The extraction flow SHALL request 2-7 place candidates from the model but SHALL not fail solely because the returned count is outside that range.

#### Scenario: Returned candidates fewer than 2 or greater than 7
- **WHEN** the model response is structurally valid but contains candidate count outside 2-7
- **THEN** the command SHALL keep the run successful
- **AND** it SHALL persist the raw response artifact for inspection

### Requirement: Response must be parseable into minimal extraction contract
The extraction flow SHALL validate model output against a minimal structured contract that includes `place` and `sentenceStartTimestamp`.

#### Scenario: Response parses to required fields
- **WHEN** each extracted candidate includes `place` and `sentenceStartTimestamp`
- **THEN** the command SHALL treat the extraction response as valid

#### Scenario: Response cannot be parsed into required fields
- **WHEN** the model output is non-JSON or missing required fields for extraction items
- **THEN** the command SHALL fail with a parse/validation error
- **AND** it SHALL preserve the raw model response artifact for debugging
