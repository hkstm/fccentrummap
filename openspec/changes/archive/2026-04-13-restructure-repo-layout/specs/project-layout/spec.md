## ADDED Requirements

### Requirement: Repository has top-level subsystem boundaries
The repository SHALL organize the scraper pipeline, frontend area, durable docs, and generated data as separate top-level concerns.

#### Scenario: Repository layout
- **WHEN** a contributor inspects the repository root
- **THEN** they SHALL find dedicated top-level locations for `scraper/`, `viz/`, `docs/`, and `data/`

### Requirement: Go module root lives under `scraper/`
The Go scraper pipeline SHALL use `scraper/` as its module root.

#### Scenario: Building the scraper
- **WHEN** a contributor runs Go build or test commands for the scraper
- **THEN** those commands SHALL execute against the module rooted at `scraper/`

### Requirement: Generated artifacts use stable locations
Generated pipeline artifacts SHALL use durable locations aligned to their roles.

#### Scenario: SQLite database location
- **WHEN** the scraper writes its SQLite database
- **THEN** the database SHALL be stored under `data/`

#### Scenario: Frontend export location
- **WHEN** the export step writes frontend-consumable data
- **THEN** it SHALL write the generated JSON to `viz/public/data/spots.json`
