## 1. Define canonical package map and migration rules

- [x] 1.1 Inventory current service/business-logic files across `scraper/internal/pipeline/*`, `scraper/internal/scraper/*`, and related capability packages.
- [x] 1.2 Create a canonical target package map that classifies each file as stage-owned, capability-owned, or cross-stage primitive.
- [x] 1.3 Document and codify placement rules (including `pipeline/common` constraints and `internal/scraper` deprecation rule) in architecture/development docs.

## 2. Introduce/normalize explicit capability packages

- [x] 2.1 Create or normalize explicit capability packages for reusable logic (audio, transcription, geocoder, content/article fetching, extraction as needed).
- [x] 2.2 Move reusable non-stage-specific business logic from legacy locations into the canonical capability packages.
- [x] 2.3 Add temporary compatibility wrappers/aliases only where necessary to keep migration increments buildable.

## 3. Migrate stage package ownership boundaries

- [x] 3.1 Ensure each parity-critical stage package under `scraper/internal/pipeline/<stage>` owns its stage DTOs, orchestration, and ports.
- [x] 3.2 Remove new and existing direct dependencies from stage packages to deprecated catch-all `internal/scraper` paths by switching to explicit capability boundaries.
- [x] 3.3 Restrict `pipeline/common` to cross-stage primitives and move any domain-specific logic to stage or capability packages.

## 4. Update imports, tests, and behavior-parity checks

- [x] 4.1 Update Go imports and package references after file moves; remove stale references to deprecated package paths.
- [x] 4.2 Update and run stage-focused tests covering parity-critical flows to verify unchanged success/failure semantics.
- [x] 4.3 Run full scraper test suite and targeted CLI smoke checks to confirm no user-facing behavior or artifact contract regressions.

## 5. Complete deprecation and documentation rollout

- [x] 5.1 Remove temporary compatibility wrappers once all call sites are migrated and stable.
- [x] 5.2 Finalize docs that describe canonical internal service locations and ownership decision examples for contributors.
- [x] 5.3 Add lightweight guardrails (lint/checklist/contributor guidance) to prevent new business logic from being introduced in deprecated catch-all package locations.
