# static-site-data-dump Specification

## Purpose
TBD - created by archiving change export-scraping-data-to-json. Update Purpose after archive.
## Requirements
### Requirement: Static data dump document for site consumption
The system SHALL generate a single JSON document containing top-level `spots` and `presenters` collections for static site consumption.

#### Scenario: Export produces top-level collections
- **WHEN** the user runs the export command against a valid SQLite dataset
- **THEN** the output file MUST contain a JSON object with `spots` and `presenters` keys

### Requirement: Spot records include required frontend fields
Each entry in `spots` MUST include `spotId`, `placeId`, `spotName`, `presenterName`, and `youtubeLink`. `spotId` SHALL be a stable identifier suitable for spot-specific links and maintainer correction workflows.

#### Scenario: Spot entry contains required fields
- **WHEN** a spot exists in the source dataset
- **THEN** the exported spot record MUST include `spotId`, `placeId`, `spotName`, `presenterName`, and `youtubeLink`

#### Scenario: Spot identifier remains stable across corrections
- **WHEN** a spot has corrected `spotName`, `placeId`, coordinates, or timestamp values
- **THEN** the exported `spotId` SHALL remain the same stable identifier for that source spot

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

