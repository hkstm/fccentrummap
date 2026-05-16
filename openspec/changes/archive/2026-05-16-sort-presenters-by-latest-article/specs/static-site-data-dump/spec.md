## MODIFIED Requirements

### Requirement: Presenter list is exported from database values as-is
The `presenters` collection SHALL contain presenter names taken directly from stored database values without normalization or canonicalization in v1, and SHALL order presenters by each presenter's latest associated article publication time descending without adding publication-time fields to the exported presenter objects.

#### Scenario: Presenter value preserves stored formatting
- **WHEN** a presenter name is stored in the database
- **THEN** the corresponding exported `presenterName` MUST match the stored value exactly

#### Scenario: Presenters are ordered by latest article publication time
- **WHEN** multiple presenters have associated articles with stored publication times
- **THEN** the exported `presenters` array SHALL order those presenters by their latest associated article publication time from newest to oldest
- **AND** each presenter object SHALL contain the existing `presenterName` field without exposing the publication time

#### Scenario: Presenter ordering uses deterministic tie-breakers
- **WHEN** two presenters have the same latest publication time
- **THEN** the export SHALL order them deterministically by `presenterName` ascending within that ordering group

#### Scenario: Missing publication time blocks export
- **WHEN** an exported presenter is associated with an article that has no stored publication time
- **THEN** the export SHALL fail with an actionable error instead of silently falling back to alphabetical ordering
