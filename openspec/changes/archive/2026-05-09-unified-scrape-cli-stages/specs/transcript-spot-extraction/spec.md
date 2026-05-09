## ADDED Requirements

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
