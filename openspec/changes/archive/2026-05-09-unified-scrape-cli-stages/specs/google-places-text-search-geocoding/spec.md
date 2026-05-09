## ADDED Requirements

### Requirement: Geocode stage is file-mode only in this change
In this change scope, geocode stage execution SHALL support file-mode input/output handoff and SHALL reject sqlite mode.

#### Scenario: Geocode invoked in sqlite mode
- **WHEN** a user runs `geocode-spots` with `--io sqlite` (explicit or default)
- **THEN** the command SHALL fail with non-zero status
- **AND** it SHALL explain that sqlite persistence integration is deferred and file mode is required

#### Scenario: Geocode invoked in file mode
- **WHEN** a user runs `geocode-spots --io file --in <path>`
- **THEN** the stage SHALL process the explicit input artifact
- **AND** it SHALL emit deterministic geocode output artifact(s)
