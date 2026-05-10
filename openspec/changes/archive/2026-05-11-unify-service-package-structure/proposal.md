## Why

The current scraper internals mix legacy domain/service packages (for example `internal/scraper`) with newer stage-first pipeline packages, which makes ownership boundaries unclear and increases maintenance cost. We need a single, explicit package organization model now to keep future pipeline work consistent and prevent architecture drift.

## What Changes

- Define and enforce a unified internal package structure for service/business logic used by scraper stages.
- Standardize where stage orchestration code lives versus where reusable domain services live.
- Deprecate ambiguous legacy service package placement (notably catch-all `internal/scraper`) in favor of explicit capability packages and/or stage-owned packages.
- Define migration and compatibility expectations so existing stage behavior remains unchanged during package moves.
- Add architecture/documentation requirements that prohibit introducing new business logic into deprecated legacy locations.

## Capabilities

### New Capabilities
- `service-package-organization`: Defines the canonical package layout and ownership boundaries for stage services, shared domain services, and legacy-package deprecation rules.

### Modified Capabilities
- `pipeline-layered-architecture`: Tighten requirements so stage packages and service boundaries align with the unified package organization.
- `project-layout`: Update required repository/package layout conventions to reflect the new canonical service organization.

## Impact

- Affected code: `scraper/internal/pipeline/*`, `scraper/internal/scraper/*`, and related service-oriented packages such as geocoding/transcription/extraction helpers.
- Affected docs/specs: architecture and development guidance describing CLI → service → adapter layering and package ownership.
- Runtime/API impact: no intended end-user CLI behavior changes; this is an internal architecture/maintainability change.
- Testing impact: existing tests will need import-path updates and regression validation to ensure no behavior change during moves.
