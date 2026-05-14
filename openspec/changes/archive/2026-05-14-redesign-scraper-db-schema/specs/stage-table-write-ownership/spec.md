## ADDED Requirements

### Requirement: Each persistence table SHALL have exactly one writer stage
The pipeline SHALL assign each persistent table to exactly one designated writer stage. Stages MAY write multiple tables, but no table MAY be written by more than one stage.

#### Scenario: Ownership mapping is defined for all pipeline tables
- **WHEN** maintainers inspect pipeline storage design and adapters
- **THEN** each table SHALL have one documented writer stage
- **AND** no table SHALL appear under multiple writer stages

#### Scenario: Stage implementation attempts cross-owned writes
- **WHEN** a stage attempts to write a table owned by another stage
- **THEN** the change SHALL be rejected by review and contract tests
- **AND** ownership documentation SHALL be updated only when intentionally reassigning single-writer ownership

### Requirement: Stage write ownership SHALL remain stable under retries
The pipeline SHALL preserve deterministic outcomes under re-execution by ensuring each stage writes only its owned tables using idempotent insert/upsert behavior.

#### Scenario: Stage rerun after partial failure
- **WHEN** a stage is rerun for already-processed source rows
- **THEN** writes to that stage’s owned tables SHALL be idempotent
- **AND** rows in tables owned by other stages SHALL remain unchanged
