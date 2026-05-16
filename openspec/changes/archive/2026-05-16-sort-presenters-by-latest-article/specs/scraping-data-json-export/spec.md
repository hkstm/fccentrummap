## MODIFIED Requirements

### Requirement: Deterministic ordering in output
The system SHALL produce deterministic ordering of exported arrays to make output stable across runs with unchanged source data. Exported spots SHALL retain stable ordering by their existing deterministic spot fields, and exported presenters SHALL be ordered by latest associated article publication time descending with `presenterName` ascending as the tie-breaker.

#### Scenario: Repeated export with unchanged data
- **WHEN** the export is run multiple times against unchanged source data
- **THEN** the resulting JSON content MUST preserve consistent array ordering

#### Scenario: Presenter ordering uses stored article publish metadata
- **WHEN** exported presenters have associated articles with stored publication times
- **THEN** the export SHALL sort the `presenters` array by each presenter's latest associated article publication time from newest to oldest

#### Scenario: Presenter ordering uses deterministic tie-breakers
- **WHEN** multiple exported presenters have the same latest associated article publication time
- **THEN** those presenters SHALL sort by `presenterName` ascending

#### Scenario: Export fails without required publish metadata
- **WHEN** one or more exported presenters are associated with exportable articles that have no stored publication time
- **THEN** the export SHALL fail with an actionable error instead of producing JSON with fallback ordering
