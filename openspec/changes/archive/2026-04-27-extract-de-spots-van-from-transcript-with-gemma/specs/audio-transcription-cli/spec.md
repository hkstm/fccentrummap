## MODIFIED Requirements

### Requirement: CLI exports stored artifacts to the data directory
The system SHALL provide export commands for source audio and transcription JSON artifacts by explicit IDs, and SHALL provide a dry-run extraction export path that writes transcript, prompt, and model-response artifacts for inspection.

#### Scenario: Export audio by audio source ID
- **WHEN** the user requests audio export with `--audio-source-id <id>`
- **THEN** the CLI SHALL write the audio payload to `data/` with a deterministic filename containing that ID
- **AND** it SHALL preserve the correct file extension based on stored format/metadata

#### Scenario: Export transcription JSON by transcription ID
- **WHEN** the user requests transcription export with `--transcription-id <id>`
- **THEN** the CLI SHALL write the canonical JSON payload to `data/` with a deterministic filename containing that ID
- **AND** it SHALL produce valid UTF-8 JSON text

#### Scenario: Run transcript-to-spot extraction dry-run
- **WHEN** the user runs the extraction dry-run command
- **THEN** the CLI SHALL write a transcript artifact, the composed Dutch prompt artifact, and raw model response artifact to `data/`
- **AND** it SHALL not persist extracted place results to DB in this phase
