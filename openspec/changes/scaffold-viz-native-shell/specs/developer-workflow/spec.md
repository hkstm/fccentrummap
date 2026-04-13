## MODIFIED Requirements

### Requirement: Root Makefile is the canonical workflow entrypoint
The repository SHALL provide a top-level `Makefile` as the canonical entrypoint for common developer workflows across scraper and frontend subsystems.

#### Scenario: Root workflow commands
- **WHEN** a contributor needs to scrape data, export frontend JSON, build, run checks, or work on the frontend shell
- **THEN** the repository SHALL provide canonical root commands for those workflows

### Requirement: Root commands delegate to subsystem paths
Repo-level commands SHALL encapsulate subsystem-specific paths so contributors can work from the repository root.

#### Scenario: Frontend setup delegation
- **WHEN** a contributor installs or runs the frontend workspace from the repository root
- **THEN** the root workflow commands SHALL target `viz/` without requiring manual directory changes

#### Scenario: Native shell workflow delegation
- **WHEN** a contributor runs a native-shell-oriented workflow from the repository root
- **THEN** the root workflow commands SHALL target the Tauri scaffold under `viz/` without implying that the shell already implements product features
