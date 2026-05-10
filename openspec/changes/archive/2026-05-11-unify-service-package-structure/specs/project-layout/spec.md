## MODIFIED Requirements

### Requirement: Repository has explicit top-level subsystem boundaries
The repository SHALL organize durable concerns into top-level locations for the Go scraper pipeline, generated data, OpenSpec artifacts, documentation, and the future frontend area. Within the scraper module, internal service code SHALL also follow a canonical layout separating stage orchestration packages, capability service packages, and shared primitives.

#### Scenario: Repository layout
- **WHEN** a contributor inspects the repository root
- **THEN** they SHALL find `scraper/`, `data/`, `openspec/`, `docs/`, and `viz/` as the primary project areas
- **AND** they SHALL be able to identify canonical internal service ownership boundaries within `scraper/internal`

## ADDED Requirements

### Requirement: Scraper internal layout SHALL expose canonical service locations
The scraper module SHALL maintain discoverable internal package locations where stage orchestration lives under `scraper/internal/pipeline/<stage>`, reusable capability services live under explicitly named capability packages, and cross-stage primitives live under `scraper/internal/pipeline/common`.

#### Scenario: Locating service implementations for maintenance
- **WHEN** a maintainer needs to modify or add service behavior
- **THEN** they SHALL be able to determine the target package from ownership rules without relying on legacy catch-all package conventions
- **AND** new business logic SHALL not require introducing additional generic catch-all service packages
