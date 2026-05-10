## 1. Dependency and CLI scaffold

- [x] 1.1 Add `github.com/urfave/cli/v3` to `scraper/go.mod` and tidy modules.
- [x] 1.2 Create urfave/cli v3 app bootstrap for the scraper CLI entrypoint.
- [x] 1.3 Define the unified command tree with subcommands: `init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, `export-data`.

## 2. Port command/flag wiring to urfave/cli v3

- [x] 2.1 Implement thin adapters from each urfave subcommand action to existing stage handlers.
- [x] 2.2 Port global/shared flags and per-stage flags from stdlib `flag.FlagSet` to urfave/cli v3 flag definitions.
- [x] 2.3 Remove non-essential legacy invocation compatibility paths and aliases during porting.

## 3. Preserve required stage behavior contracts

- [x] 3.1 Verify unsupported stage/mode combinations still fail with non-zero status and actionable guidance (including `geocode-spots` default sqlite mode behavior).
- [x] 3.2 Verify `transcribe-audio` unified-stage invocation still persists canonical transcription rows in SQLite.
- [x] 3.3 Verify file-mode transcription still accepts explicit `--in` and produces deterministic output artifact naming.

## 4. Update docs and remove old wiring

- [x] 4.1 Update in-repo CLI docs/examples to the idiomatic urfave/cli v3 interface and accepted breaking invocation changes.
- [x] 4.2 Remove obsolete stdlib `flag` command-wiring code paths from CLI entrypoints after urfave wiring is active.
- [x] 4.3 Run a final command usability smoke pass for main scrape/transcription flows and capture any doc fixes needed.
