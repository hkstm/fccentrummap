## ADDED Requirements

### Requirement: Dry-run command exports a single transcript for inspection
The CLI dry-run workflow SHALL export the selected transcript payload from SQLite to a local file before model invocation.

#### Scenario: Run with explicit transcript selection
- **WHEN** the user runs the extraction dry-run command with a transcript identifier
- **THEN** the command SHALL read that transcript from storage
- **AND** it SHALL write an exported transcript artifact file under `data/`

### Requirement: Dry-run command writes composed Dutch prompt to disk
The dry-run workflow SHALL persist the exact prompt sent to the model as a text artifact.

#### Scenario: Prompt artifact creation
- **WHEN** the command composes the extraction prompt
- **THEN** it SHALL write the full prompt text to a deterministic file under `data/`
- **AND** the artifact SHALL be inspectable without rerunning prompt construction

### Requirement: Dry-run command writes raw model response to disk
The dry-run workflow SHALL persist the raw model API response payload without normalization.

#### Scenario: Raw response artifact creation
- **WHEN** the model call completes
- **THEN** the command SHALL write the raw response body to a deterministic file under `data/`
- **AND** this file SHALL be suitable for manual inspection and later parser iteration

### Requirement: Dry-run mode does not persist extraction output to database
The dry-run workflow SHALL not write extracted place results back to DB tables.

#### Scenario: Successful dry-run completion
- **WHEN** transcript export, prompt creation, and model response artifact writing succeed
- **THEN** the command SHALL complete without inserting or updating extracted-place records in SQLite
