## Purpose

Define the canonical repository layout, subsystem boundaries, and default artifact locations.

## Requirements

### Requirement: Repository has explicit top-level subsystem boundaries
The repository SHALL organize durable concerns into top-level locations for the Go scraper pipeline, generated data, OpenSpec artifacts, documentation, and the future frontend area.

#### Scenario: Repository layout
- **WHEN** a contributor inspects the repository root
- **THEN** they SHALL find `scraper/`, `data/`, `openspec/`, `docs/`, and `viz/` as the primary project areas

### Requirement: Go module root lives under `scraper/`
The Go codebase SHALL use `scraper/` as its module root rather than the repository root.

#### Scenario: Running Go commands
- **WHEN** a contributor runs Go build or test commands for the scraper pipeline
- **THEN** those commands SHALL execute from the module rooted at `scraper/`

### Requirement: Generated artifacts use stable locations
Generated artifacts SHALL live in locations aligned to their roles.

#### Scenario: SQLite database location
- **WHEN** the scraper writes or reads its SQLite database
- **THEN** the database path SHALL be `data/spots.db` by default

#### Scenario: Frontend export location
- **WHEN** the export step writes frontend-consumable data
- **THEN** it SHALL write JSON to `viz/public/data/spots.json`
