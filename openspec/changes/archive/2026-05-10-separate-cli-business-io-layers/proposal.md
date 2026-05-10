## Why

The current scraper pipeline mixes CLI routing, stage orchestration, and persistence concerns, which makes file-mode behavior uneven and harder to evolve safely. We need a clear layering model now so stage logic can be reused consistently across SQLite and file-backed flows without duplicating behavior.

## What Changes

- Introduce explicit layering boundaries: **CLI layer → business logic layer → I/O adapters**.
- Define business-layer stage contracts that are independent of concrete storage engines.
- Add adapter interfaces for stage input/output so stages can run against SQLite or file artifacts with equivalent semantics where supported.
- Standardize file-mode stage behavior from passthrough scaffolding to contract-driven read/transform/write flows.
- Preserve existing SQLite-first behavior and validation semantics while making backend-specific constraints explicit.
- **BREAKING**: Stage implementations may reject legacy invocation/edge behaviors that bypass the new contracts.

## Capabilities

### New Capabilities
- `pipeline-layered-architecture`: Defines required separation between CLI command wiring, business-stage orchestration, and pluggable I/O adapters.

### Modified Capabilities
- `unified-scrape-cli`: Update requirements so stage commands delegate into backend-agnostic business services rather than embedding persistence-specific flows.
- `stage-artifact-file-io`: Upgrade requirements from simple artifact passthrough to typed stage handoff contracts with deterministic input/output semantics.
- `sqlite-storage`: Clarify SQLite adapter responsibilities and integrity guarantees within the layered architecture.

## Impact

- Affected code:
  - `scraper/cmd/scrape/main.go`
  - stage orchestration in `scraper/internal/scraper/*`, `scraper/internal/extractspots/*`
  - repository access boundaries in `scraper/internal/repository/*`
  - artifact I/O helpers in `scraper/internal/cliutil/*`
- Affected interfaces:
  - Stage invocation contracts and internal service boundaries
  - File-mode artifact schemas and stage handoff expectations
- Affected docs/specs:
  - New capability spec for layering
  - Delta specs for unified CLI, file I/O stage contracts, and SQLite role definition
