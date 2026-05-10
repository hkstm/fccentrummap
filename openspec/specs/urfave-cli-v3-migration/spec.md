# urfave-cli-v3-migration Specification

## Purpose
TBD - created by archiving change migrate-urfave-cli-v3. Update Purpose after archive.
## Requirements
### Requirement: CLI entrypoints SHALL use urfave/cli v3 command model
The system SHALL implement scraper-related CLI entrypoints using `urfave/cli/v3` commands, flags, and action handlers instead of stdlib `flag.FlagSet` command parsing.

#### Scenario: Build command tree from urfave app definition
- **WHEN** a user runs the scraper CLI entrypoint
- **THEN** the CLI SHALL resolve subcommands and flags through urfave/cli v3 command definitions
- **AND** command routing SHALL be handled by urfave actions

### Requirement: Migration SHALL prefer idiomatic urfave behavior over legacy compatibility
The system SHALL prioritize idiomatic urfave/cli v3 invocation and error/help behavior and SHALL NOT preserve deprecated compatibility aliases or shims unless strictly required for core operation.

#### Scenario: Legacy-only compatibility path is not retained
- **WHEN** an old invocation pattern exists only to mirror prior stdlib-flag behavior
- **THEN** the migrated CLI SHALL be allowed to remove that pattern
- **AND** project docs SHALL describe the supported urfave/cli v3 usage

### Requirement: In-repo CLI documentation SHALL be updated for migrated interface
The system SHALL update repository documentation and examples so CLI usage matches the urfave/cli v3 command and flag interface.

#### Scenario: Maintainer follows docs after migration
- **WHEN** a maintainer uses documented CLI examples from this repository
- **THEN** the examples SHALL match current urfave/cli v3 invocation patterns
- **AND** no obsolete stdlib-flag command wiring instructions SHALL remain

