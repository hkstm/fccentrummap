## MODIFIED Requirements

### Requirement: Repository has top-level subsystem boundaries
The repository SHALL organize durable concerns into top-level locations for the Go scraper pipeline, generated data, OpenSpec artifacts, documentation, and the frontend workspace.

#### Scenario: Repository layout
- **WHEN** a contributor inspects the repository root
- **THEN** they SHALL find `scraper/`, `data/`, `openspec/`, `docs/`, and `viz/` as the primary project areas
- **AND** `viz/` MAY contain both the web frontend workspace and a native-shell scaffold for that frontend
