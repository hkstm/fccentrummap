## ADDED Requirements

### Requirement: Extraction output includes a single presenter name
The extraction flow SHALL include a top-level `presenter_name` field in the structured model output for each extraction run.

#### Scenario: Presenter name is present in model output
- **WHEN** the model returns a presenter identity with extracted places
- **THEN** the parser SHALL map that value to `presenter_name`
- **AND** persisted extraction data SHALL include `presenter_name`

#### Scenario: Presenter name is missing in model output
- **WHEN** the model output omits presenter identity
- **THEN** the parser SHALL set `presenter_name` to null/empty according to storage conventions
- **AND** the extraction run SHALL remain valid if place/timestamp requirements are satisfied

### Requirement: Presenter attribution applies to the full extraction run
The system SHALL treat `presenter_name` as run-level metadata shared by all extracted places from that run.

#### Scenario: Multiple places in one extraction result
- **WHEN** the model output contains multiple extracted places
- **THEN** the system SHALL store one `presenter_name` value for the run
- **AND** consumers SHALL interpret that presenter as the person describing the extracted spots
