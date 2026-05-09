## ADDED Requirements

### Requirement: Stage-mode support matrix is enforced by command-layer validation
The system SHALL enforce declared stage/mode support and reject unsupported combinations before processing.

#### Scenario: Unsupported stage/mode requested
- **WHEN** a stage is requested with an unsupported I/O mode
- **THEN** validation SHALL fail before any data mutation
- **AND** the command SHALL return a non-zero error with guidance

### Requirement: Current change defers new geocode-to-final-table SQLite writes
This change SHALL NOT require introducing new SQLite write paths from geocode stage into final export tables.

#### Scenario: Geocode stage executed in current scope
- **WHEN** geocode stage is executed in this change scope
- **THEN** SQLite final-table writes from geocode SHALL be treated as deferred
- **AND** file artifact handoff SHALL remain the temporary supported contract
