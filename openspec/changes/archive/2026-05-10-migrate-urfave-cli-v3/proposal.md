## Why

The CLI is currently implemented with Go’s standard `flag` package, which makes subcommand ergonomics, help customization, and command composition more manual as the CLI grows. Migrating now to `urfave/cli/v3` gives us a clearer command framework before the interface expands further.

## What Changes

- Introduce `urfave/cli/v3` as the command framework across the Go CLI entrypoints, replacing the current stdlib `flag`-set wiring.
- Refactor command/flag/action wiring to match v3 APIs while preserving existing user-visible behavior where possible.
- Validate stage/subcommand invocation, exit behavior, and help output for unified scrape and transcription flows after migration.
- Update developer-facing CLI usage docs and examples to match any v3-driven invocation/help differences.
- **BREAKING**: If urfave/cli v3 requires incompatible flag parsing/help formatting or command wiring changes, CLI output and invocation edge cases may change.

## Capabilities

### New Capabilities
- `urfave-cli-v3-migration`: Establishes project-wide CLI framework behavior and compatibility expectations after upgrading to urfave/cli v3.

### Modified Capabilities
- `unified-scrape-cli`: Update command/subcommand and option requirements to reflect urfave/cli v3 wiring and behavior.
- `audio-transcription-cli`: Update transcription CLI command/flag requirements to reflect urfave/cli v3 behavior.

## Impact

- Affected code: Go CLI entrypoints, command registration, flag definitions, action handlers, and CLI tests.
- Affected dependencies: add `github.com/urfave/cli/v3` (and any related transitive changes), while retiring direct stdlib-flag command wiring patterns in app entrypoints.
- Affected interfaces: CLI invocation semantics, help text/formatting, and error messaging paths.
- Affected docs/workflows: README or command usage docs used by developers running scrape/transcription stages.
