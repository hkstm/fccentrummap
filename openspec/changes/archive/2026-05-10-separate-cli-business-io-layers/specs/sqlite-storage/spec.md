## ADDED Requirements

### Requirement: SQLite backend SHALL act as an adapter behind stage ports
SQLite-backed persistence SHALL be accessed through stage adapter implementations that satisfy stage port contracts used by business services.

#### Scenario: Stage service executes in sqlite mode
- **WHEN** a parity-critical stage is run with `--io sqlite`
- **THEN** the stage service SHALL call SQLite adapter implementations through stage ports
- **AND** the service layer SHALL remain independent of concrete SQLite repository types

### Requirement: SQLite adapter SHALL preserve integrity constraints as canonical guarantees
SQLite adapter behavior SHALL preserve canonical integrity guarantees defined by schema constraints and transactional repository behavior.

#### Scenario: SQLite adapter persists linked stage data
- **WHEN** SQLite adapter writes related records for a stage flow
- **THEN** foreign-key and uniqueness guarantees SHALL be enforced by SQLite schema constraints
- **AND** constraint violations SHALL surface as non-zero stage failures with actionable errors
