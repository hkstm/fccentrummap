## ADDED Requirements

### Requirement: SQLite schema stores spot corrections
The repository schema SHALL include durable storage for manual spot corrections keyed by stable spot identifier.

#### Scenario: Fresh database initialization includes corrections table
- **WHEN** the repository initializes a fresh SQLite database
- **THEN** it SHALL create a `spot_corrections` table or equivalent correction storage with stable spot identifier, nullable spot name override, nullable place ID override, nullable corrected latitude/longitude, nullable timestamp override, hidden flag, and audit timestamps

#### Scenario: Existing database initialization adds corrections table
- **WHEN** the repository initializes an existing compatible database without correction storage
- **THEN** it SHALL create the missing correction storage idempotently without dropping existing source data

### Requirement: Repository upserts spot corrections
The repository SHALL expose operations to save and retrieve the current correction for a stable spot identifier.

#### Scenario: Insert correction
- **WHEN** no correction exists for a stable spot identifier
- **THEN** saving a correction SHALL insert one correction row for that identifier

#### Scenario: Update correction
- **WHEN** a correction already exists for a stable spot identifier
- **THEN** saving a correction SHALL update that existing correction row
- **AND** it SHALL preserve one active correction per stable spot identifier

### Requirement: Export query includes correction data
The repository SHALL make stored spot corrections available to export assembly without mutating source spot records.

#### Scenario: Exporting corrected and uncorrected spots
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return each exportable spot with its stable spot identifier and any stored correction values
- **AND** uncorrected spots SHALL remain exportable with their original source values
