## Purpose
Define the transcript-to-spot extraction contract, including two-pass timestamp refinement, validation, persistence, and operator-controlled failure handling.

## Requirements

### Requirement: Extraction input uses sentence-level transcript units with timestamps
The extraction flow SHALL run in two passes. Pass 1 SHALL build model input from cleaned article text and sentence-level transcript units with sentence start timestamps. Pass 2 SHALL run as a single batched refinement call using transcript sentence units plus pass-1 place/timestamp outputs, without requiring cleaned article text.

#### Scenario: Cleaned article and sentence transcript are available
- **WHEN** the selected article has cleaned article text and transcript sentence-level units with `text` and `start` fields
- **THEN** the command SHALL execute pass 1 to extract places with initial timestamps
- **AND** it SHALL execute pass 2 in one batched call for all pass-1 places using transcript context plus pass-1 timestamps
- **AND** pass 2 input SHALL not require cleaned article text

#### Scenario: Sentence-level transcript missing
- **WHEN** sentence-level transcript units are missing and only full-text blobs or word-level token timings are present
- **THEN** the command SHALL fail with a clear validation error
- **AND** it SHALL not send pass-1 or pass-2 model requests

### Requirement: Prompt requires Dutch extraction focused on Amsterdam region
The extraction prompts SHALL be written in Dutch. Pass 1 SHALL instruct the model to extract places in Amsterdam or nearby surroundings using cleaned article text as primary place source and transcript sentence units for evidence/timing. Pass 2 SHALL instruct the model to refine timestamps for the already-extracted places using transcript evidence only.

#### Scenario: Prompt composition for two-pass extraction run
- **WHEN** the command composes prompts for pass 1 and pass 2
- **THEN** pass-1 prompt SHALL instruct the model to find "De spots van" place mentions
- **AND** pass-1 prompt SHALL constrain candidate places to Amsterdam or nearby surroundings
- **AND** pass-2 prompt SHALL request earliest logical timestamp anchors for the same extracted places from pass 1

### Requirement: Response must be parseable into timestamp-refined extraction contract
The extraction flow SHALL validate model output so each accepted place has a `place`, an `originalSentenceStartTimestamp` from pass 1, and a `refinedSentenceStartTimestamp` from pass 2. The refined timestamp MAY equal the original timestamp when no earlier logical anchor exists.

#### Scenario: Response parses to required fields with earlier refinement
- **WHEN** pass 1 returns `place` plus initial timestamp and pass 2 returns a valid earlier-or-equal refined timestamp for that place
- **THEN** the command SHALL treat the extraction response as valid
- **AND** it SHALL persist both original and refined timestamps for that place

#### Scenario: Response parses to required fields with no-op refinement
- **WHEN** pass 2 determines the original timestamp is already the earliest logical anchor
- **THEN** the command SHALL keep the original timestamp unchanged as the refined timestamp
- **AND** it SHALL treat the result as valid

#### Scenario: Refinement output violates ordering constraints
- **WHEN** a pass-2 refined timestamp is later than its pass-1 original timestamp or is not strictly greater than the previous place’s pass-1 original timestamp in pass-1 order
- **THEN** the command SHALL reject that refined timestamp update
- **AND** it SHALL retain the original timestamp for that place

### Requirement: Unified extract-spots stage keeps SQLite record and artifact outputs
The extract-spots stage SHALL preserve SQLite persistence of extraction records and also support deterministic artifact outputs for inspection/handoff.

#### Scenario: Extract-spots in sqlite mode
- **WHEN** extract-spots runs in sqlite mode
- **THEN** it SHALL read required inputs from SQLite
- **AND** it SHALL persist extraction record output in SQLite
- **AND** it MAY additionally write deterministic stage artifacts for prompts/responses

### Requirement: Expensive-stage retry policy is operator-controlled
The extraction stage SHALL not auto-retry expensive model calls after failure.

#### Scenario: Model call fails
- **WHEN** an extract-spots model request fails
- **THEN** the stage SHALL fail explicitly
- **AND** subsequent retries SHALL require an explicit new operator invocation
