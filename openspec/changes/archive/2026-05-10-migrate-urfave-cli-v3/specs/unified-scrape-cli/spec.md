## MODIFIED Requirements

### Requirement: Unified scrape CLI provides stage subcommands
The system SHALL provide a single urfave/cli v3 entrypoint for pipeline execution with subcommands: `init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, and `export-data`.

#### Scenario: List and run unified stages
- **WHEN** a user invokes the unified scrape CLI
- **THEN** the listed stage subcommands SHALL be available under that single entrypoint

### Requirement: Unsupported stage/mode combinations fail explicitly
The unified CLI SHALL return a non-zero error with actionable guidance when a stage is invoked in an unsupported I/O mode.

#### Scenario: Unsupported geocode sqlite mode
- **WHEN** a user runs `geocode-spots` with default `--io sqlite`
- **THEN** the CLI SHALL fail explicitly
- **AND** it SHALL instruct the user to use the currently supported mode (file mode with explicit `--in`)

## ADDED Requirements

### Requirement: Unified scrape CLI MAY adopt idiomatic urfave/cli v3 flag conventions
The unified scrape CLI SHALL allow urfave/cli v3-native flag handling and command ergonomics, and it SHALL NOT be required to preserve exact legacy stdlib-flag invocation edge cases.

#### Scenario: Invocation differs from legacy parsing but remains documented
- **WHEN** a previously accepted edge-case invocation is incompatible with urfave/cli v3 conventions
- **THEN** the CLI MAY reject that invocation
- **AND** the repository documentation SHALL reflect the supported invocation form
