## ADDED Requirements

### Requirement: Root Makefile is the canonical workflow entrypoint
The repository SHALL provide a top-level `Makefile` as the canonical entrypoint for common developer workflows.

#### Scenario: Repo-level workflow
- **WHEN** a contributor needs to run a common workflow from the repository root
- **THEN** the `Makefile` SHALL provide canonical commands for scrape, export, build, and verification tasks

### Requirement: Repo-level commands delegate to subsystem locations
Repo-level workflow commands SHALL encapsulate subsystem-specific paths so contributors can work from the repository root.

#### Scenario: Scraper command delegation
- **WHEN** a contributor runs a scraper-related root command
- **THEN** that command SHALL invoke the relevant `scraper/` workflow without requiring the contributor to manually change directories

#### Scenario: Frontend command delegation
- **WHEN** a contributor runs a frontend-related root command
- **THEN** that command SHALL target the `viz/` area without assuming the frontend lives at the repository root
