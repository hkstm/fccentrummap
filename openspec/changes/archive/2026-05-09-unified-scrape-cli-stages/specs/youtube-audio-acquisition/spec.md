## ADDED Requirements

### Requirement: Audio acquisition stage supports unified I/O contract
The audio acquisition stage SHALL support unified CLI I/O mode behavior with SQLite as default and explicit failure for unsupported combinations.

#### Scenario: Acquire audio in sqlite mode
- **WHEN** the stage runs in default sqlite mode
- **THEN** it SHALL read candidate records from SQLite and persist acquired audio metadata/blob to SQLite

#### Scenario: Unsupported invocation mode
- **WHEN** the stage is invoked in a mode not implemented for that command
- **THEN** it SHALL fail with a non-zero status and actionable guidance
