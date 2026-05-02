## MODIFIED Requirements

### Requirement: Dry-run command writes composed Dutch prompt to disk
The dry-run workflow SHALL persist the exact prompts sent to the model as text artifacts for both extraction passes.

#### Scenario: Prompt artifact creation for two-pass flow
- **WHEN** the command composes pass-1 and pass-2 prompts
- **THEN** it SHALL write the full pass-1 prompt text to a deterministic file under `data/`
- **AND** it SHALL write the full pass-2 prompt text to a deterministic file under `data/`
- **AND** both artifacts SHALL be inspectable without rerunning prompt construction

### Requirement: Dry-run command writes raw model response to disk
The dry-run workflow SHALL persist raw model API response payloads without normalization for both extraction passes.

#### Scenario: Raw response artifact creation for two-pass flow
- **WHEN** pass-1 and pass-2 model calls complete
- **THEN** the command SHALL write the pass-1 raw response body to a deterministic file under `data/`
- **AND** it SHALL write the pass-2 raw response body to a deterministic file under `data/`
- **AND** these files SHALL be suitable for manual inspection and later parser iteration

### Requirement: Dry-run mode does not persist extraction output to database
The dry-run workflow SHALL not create, insert, update, or delete any SQLite records, even when two-pass refinement succeeds.

#### Scenario: Successful two-pass dry-run completion
- **WHEN** transcript export, both prompt artifact writes, and both model response artifact writes succeed
- **THEN** the command SHALL complete without any SQLite writes or mutations
