## ADDED Requirements

### Requirement: Go exporter reads SQLite and writes JSON
A Go command at `scraper/cmd/export/main.go` SHALL read the SQLite database and write a JSON file containing all spots with their associated authors.

#### Scenario: Successful export
- **WHEN** the exporter is run with a valid database path and output path (`cd scraper && go run ./cmd/export -db ../data/spots.db -out ../viz/public/data/spots.json`)
- **THEN** it SHALL write a JSON file to the specified output path containing all spots and authors

#### Scenario: Database file not found
- **WHEN** the exporter is run with a non-existent database path
- **THEN** it SHALL exit with a non-zero status and print an error message

### Requirement: JSON schema contains spots with author attribution
The exported JSON SHALL contain an array of spots, each with name, address, latitude, longitude, and an array of author names. It SHALL also contain a top-level array of all unique authors.

#### Scenario: JSON structure
- **WHEN** the export completes
- **THEN** the JSON file SHALL have the structure: `{ "authors": ["name", ...], "spots": [{ "name": "", "address": "", "lat": 0.0, "lng": 0.0, "authors": ["name", ...] }, ...] }`

#### Scenario: Spot with multiple authors
- **WHEN** a spot appears in articles by multiple authors
- **THEN** the spot's `authors` array SHALL contain all associated author names

### Requirement: Exporter uses existing Go dependencies
The exporter SHALL use `modernc.org/sqlite` (already in the project) for database access. It SHALL NOT introduce new Go dependencies.

#### Scenario: No new dependencies
- **WHEN** the exporter is built
- **THEN** it SHALL compile using only dependencies already present in `go.mod`

### Requirement: Frontend loads data from static JSON
The Next.js frontend SHALL fetch the JSON file from `viz/public/data/spots.json` at build/runtime via `/data/spots.json` and use it as the sole data source for rendering markers and populating the author filter.

#### Scenario: Data loads on page load
- **WHEN** the page loads
- **THEN** the frontend SHALL fetch `/data/spots.json` and parse it to populate the map and filter

#### Scenario: JSON file missing
- **WHEN** the fetch for `/data/spots.json` fails
- **THEN** the application SHALL display an error message indicating the data could not be loaded

### Requirement: Generated JSON is gitignored
The `viz/public/data/spots.json` file SHALL be listed in `.gitignore` since it is a build artifact generated from the database.

#### Scenario: File is not tracked by git
- **WHEN** the exporter generates `viz/public/data/spots.json`
- **THEN** the file SHALL be excluded from version control via `.gitignore`
