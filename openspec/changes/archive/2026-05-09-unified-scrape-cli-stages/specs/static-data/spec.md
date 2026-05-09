## ADDED Requirements

### Requirement: Export stage parity can be validated as command smoke test
During unified-CLI scaffold phase, export-data SHALL be validatable as a command/interface smoke test even when upstream pipeline stages have not populated final export join tables.

#### Scenario: Export with no final joined rows
- **WHEN** export-data runs successfully against a database where final export join rows are absent
- **THEN** the command SHALL still succeed and write JSON output
- **AND** this result SHALL be treated as smoke-test validation of command wiring, not full data-correctness validation
