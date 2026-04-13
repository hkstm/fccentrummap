## MODIFIED Requirements

### Requirement: Go exporter reads SQLite and writes JSON
A Go command at `scraper/cmd/export/main.go` SHALL read the SQLite database and write a JSON file containing all spots with their associated authors.

#### Scenario: Successful export
- **WHEN** the exporter is run with a valid database path under `data/` and output path `viz/public/data/spots.json`
- **THEN** it SHALL write a JSON file to that output path containing all spots and authors

#### Scenario: Database file not found
- **WHEN** the exporter is run with a non-existent database path
- **THEN** it SHALL exit with a non-zero status and print an error message

### Requirement: Generated JSON is gitignored
The `viz/public/data/spots.json` file SHALL be listed in `.gitignore` since it is a build artifact generated from the database.

#### Scenario: File is not tracked by git
- **WHEN** the exporter generates `viz/public/data/spots.json`
- **THEN** the file SHALL be excluded from version control via `.gitignore`
