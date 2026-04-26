## Purpose

Define the canonical repo-level developer entrypoints and how they map to subsystem workflows.

## Requirements

### Requirement: Root Makefile is the canonical workflow entrypoint
The repository SHALL provide a top-level `Makefile` as the canonical entrypoint for common developer workflows.

#### Scenario: Root workflow commands
- **WHEN** a contributor needs to scrape data, export frontend JSON, build, or run checks
- **THEN** the repository SHALL provide `make scrape`, `make export`, `make build`, and `make check`

### Requirement: Root commands delegate to subsystem paths
Repo-level commands SHALL encapsulate subsystem-specific paths so contributors can work from the repository root.

#### Scenario: Scraper command delegation
- **WHEN** a contributor runs `make scrape`
- **THEN** the command SHALL execute the scraper CLI from `scraper/cmd/scraper` while targeting `data/spots.db`

#### Scenario: Export command delegation
- **WHEN** a contributor runs `make export`
- **THEN** the command SHALL execute the exporter from `scraper/cmd/export` and write `viz/public/data/spots.json`

#### Scenario: Verification commands
- **WHEN** a contributor runs `make check`
- **THEN** the command SHALL run the scraper test suite and build checks from the `scraper/` module

### Requirement: Conventional Commit messages are enforced with local hook setup
The repository SHALL define Conventional Commits as the commit format and provide a setup command for basic local validation.
Note: this validation is implemented via local Git hooks (`make setup-hooks` -> `.githooks/commit-msg`), can be bypassed with `--no-verify`, and is not currently enforced server-side.

#### Scenario: Contributor configures hooks
- **WHEN** a contributor runs `make setup-hooks`
- **THEN** Git SHALL be configured to use repository hooks from `.githooks`

#### Scenario: Commit message validation
- **WHEN** a contributor creates a commit using local hooks
- **THEN** the commit message SHALL follow Conventional Commits (`<type>[optional scope][!]: <description>`)
