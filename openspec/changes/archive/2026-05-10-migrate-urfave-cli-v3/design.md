## Context

The current CLI is implemented with Go `flag.FlagSet` plumbing per stage command, which has worked for initial pipeline growth but makes consistency (help formatting, shared option behavior, subcommand ergonomics, and error messaging) increasingly manual. This change introduces `urfave/cli/v3` as the command framework for scraper/transcription entrypoints while preserving existing stage intent and behavior. Existing specs that define stage names and contracts (`unified-scrape-cli`, `audio-transcription-cli`) remain the source of truth for user-visible requirements.

Constraints:
- Keep stage/subcommand coverage intact (`init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, `export-data`).
- Maintain existing non-zero failure semantics and actionable error output.
- Avoid mixing old and new command-parsing flows in the same executable after migration completion.

## Goals / Non-Goals

**Goals:**
- Replace stdlib `flag` command wiring with `urfave/cli/v3` command/action definitions.
- Preserve core stage behavior and existing domain logic (scraping/transcription/geocoding internals) while preferring idiomatic urfave/cli v3 conventions over strict CLI parity.
- Standardize flag parsing/help output and reduce per-command boilerplate.
- Clean up obsolete/legacy CLI surface that is no longer strictly required.
- Create a migration structure that makes future CLI feature additions easier and less error-prone.

**Non-Goals:**
- Rewriting scraper, transcription, extraction, or geocoding business logic.
- Introducing new pipeline stages or changing pipeline data model/storage semantics.
- Committing to exact historical help text byte-for-byte compatibility.

## Decisions

1. **Adopt urfave/cli v3 as the single command framework**
   - Rationale: unified subcommand model, typed flags, and app-level lifecycle handling reduce custom parsing code.
   - Alternative considered: keep stdlib `flag` and add helper wrappers. Rejected because complexity and inconsistency still accumulate across commands.

2. **Prioritize idiomatic urfave/cli v3 command design over strict backward compatibility**
   - Rationale: there are no external consumers outside this repository; maintainability and clarity are more valuable than preserving legacy invocation quirks.
   - Alternative considered: preserve command/flag parity as much as possible. Rejected because it carries legacy complexity into the new framework.

3. **Use thin command adapters that delegate to existing stage handlers**
   - Rationale: isolates parsing/routing changes from business logic; easier to verify parity.
   - Alternative considered: rewrite handlers around urfave context objects. Rejected due to unnecessary coupling and larger regression surface.

4. **Treat help/error formatting differences as acceptable where behavior is equivalent**
   - Rationale: urfave/cli introduces formatting differences; requirement focus is functional behavior and actionable errors.
   - Alternative considered: emulate prior formatting exactly. Rejected as high effort with limited functional value.

5. **Do not preserve deprecated aliases or transition shims unless strictly required**
   - Rationale: cleanup is preferred; removing unnecessary compatibility paths keeps the CLI simpler and more idiomatic.
   - Alternative considered: temporary alias/compatibility window. Rejected per project preference for immediate cleanup.

6. **Skip dedicated golden tests for `--help` output**
   - Rationale: help-text formatting is not a priority validation target for this migration.
   - Alternative considered: snapshot/golden tests for help output. Rejected as low-value maintenance overhead.

## Risks / Trade-offs

- **[Risk] Invocation incompatibilities for edge-case flag patterns** → **Mitigation:** document intentional breaking differences and keep in-repo scripts updated in the same change.
- **[Risk] Help output changes may surprise maintainers** → **Mitigation:** update docs/examples; do not block migration on format parity.
- **[Risk] Partial migration leaves split command frameworks** → **Mitigation:** migrate all relevant entrypoints in one cohesive change before merge.
- **[Risk] Error-exit behavior differs from previous implementation** → **Mitigation:** accept idiomatic urfave/cli v3 behavior and keep implementation/docs aligned with the new command model.

## Migration Plan

1. Add `github.com/urfave/cli/v3` dependency and scaffold app/command tree mirroring existing stage set.
2. Implement command adapters that map parsed flags/options to existing stage execution functions.
3. Port global/shared options and stage-specific options with equivalent defaults and validation semantics.
4. Validate primary command usability for unified scrape and transcription flows under idiomatic urfave/cli v3 behavior, without parity requirements for legacy exit/error semantics.
5. Update in-repo CLI usage docs/examples to reflect the new idiomatic urfave/cli v3 interface.
6. Remove deprecated stdlib-flag wiring and any non-essential compatibility aliases/shims once migration checks pass.

Rollback strategy:
- Revert migration commit(s) to restore prior stdlib-flag entrypoint wiring if blocking regressions are found.

## Open Questions

- None currently. Project direction is to prioritize idiomatic urfave/cli v3 usage, accept backward-incompatible cleanup where needed, skip help-output golden tests, and update in-repo docs/scripts accordingly.
