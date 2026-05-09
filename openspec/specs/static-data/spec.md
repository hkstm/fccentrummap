## Purpose

Define the canonical SQLite-to-JSON export contract between the scraper pipeline and the future frontend.

## Requirements

### Requirement: Go exporter reads SQLite and writes JSON
A Go command at `scraper/cmd/export/main.go` SHALL read the SQLite database and write a JSON file containing all exported spots with author attribution.

#### Scenario: Successful export
- **WHEN** the exporter is run from `scraper/` with a valid database path and output path
- **THEN** it SHALL write JSON to `../viz/public/data/spots.json`

#### Scenario: Missing database file
- **WHEN** the exporter is run with a non-existent database path
- **THEN** it SHALL exit with a non-zero status and report an error

### Requirement: Export JSON contains authors and deduplicated spots
The exported JSON SHALL contain a top-level `authors` array and a `spots` array where each spot includes `name`, `address`, `lat`, `lng`, and `authors`.

#### Scenario: JSON structure
- **WHEN** export completes successfully
- **THEN** the JSON SHALL have the shape `{ "authors": [...], "spots": [{ "name": ..., "address": ..., "lat": ..., "lng": ..., "authors": [...] }] }`

#### Scenario: Spot shared by multiple authors
- **WHEN** multiple article/author links refer to the same spot name and address
- **THEN** the exporter SHALL emit one spot entry with all associated authors

### Requirement: Generated export file is not tracked in git
The generated JSON export SHALL be treated as a build artifact.

#### Scenario: Ignored export path
- **WHEN** `viz/public/data/spots.json` is generated
- **THEN** the file SHALL be excluded from version control via `.gitignore`

### Requirement: Export stage parity can be validated as command smoke test
During unified-CLI scaffold phase, export-data SHALL be validatable as a command/interface smoke test even when upstream pipeline stages have not populated final export join tables.

#### Scenario: Export with no final joined rows
- **WHEN** export-data runs successfully against a database where final export join rows are absent
- **THEN** the command SHALL still succeed and write JSON output
- **AND** this result SHALL be treated as smoke-test validation of command wiring, not full data-correctness validation
