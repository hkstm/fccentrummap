## Context

The codebase currently mixes two organizational patterns for scraper business logic: newer stage-first pipeline packages (`scraper/internal/pipeline/<stage>`) and older catch-all/domain packages (notably `scraper/internal/scraper`, plus overlapping capability packages). This causes unclear ownership, inconsistent placement of orchestration logic, and ongoing drift as new work lands in different locations.

This change is internal-only and must preserve existing CLI behavior and stage outputs. The design must make future contributions predictable: maintainers should immediately know where to add stage orchestration, reusable capability logic, and cross-stage primitives.

## Goals / Non-Goals

**Goals:**
- Establish one canonical service package organization model for scraper internals.
- Define hard boundaries between stage orchestration packages, capability service packages, and shared primitives.
- Provide a safe migration path from legacy `internal/scraper` and overlapping package placements without changing runtime behavior.
- Enforce architectural guardrails in docs/specs so new code does not reintroduce ambiguity.

**Non-Goals:**
- No feature-level changes to scrape pipeline behavior, outputs, or CLI flags.
- No backend/storage redesign beyond package placement and dependency boundaries.
- No large semantic rewrites of stage business logic during this organizational change.

## Decisions

1. **Adopt a two-tier ownership model for business logic**
   - Stage orchestration, stage DTOs, and stage ports remain stage-first in `internal/pipeline/<stage>`.
   - Reusable non-stage-specific domain logic moves to explicit capability packages (e.g., `internal/audio`, `internal/transcription`, `internal/geocoder`, `internal/contentfetch` or equivalent canonical names).
   - **Rationale:** Preserves clear pipeline boundaries while avoiding duplication across stages.
   - **Alternative considered:** Put all logic in stage packages only. Rejected due to duplicated helpers and weaker reuse semantics.

2. **Deprecate `internal/scraper` as a business-logic destination**
   - Existing code in `internal/scraper` is migrated to either stage packages or explicit capability packages based on ownership.
   - New business logic in `internal/scraper` is disallowed by convention/spec.
   - Temporary compatibility wrappers are allowed only to reduce migration risk and should be removed after import updates.
   - **Rationale:** `scraper` is too generic and has become a dumping ground.
   - **Alternative considered:** Keep `internal/scraper` and document subareas. Rejected because boundary ambiguity remains.

3. **Define placement rules based on dependency direction**
   - Stage-specific orchestration and contracts that depend on stage ports belong in `internal/pipeline/<stage>`.
   - Capability logic independent of CLI/stage orchestration belongs in capability packages.
   - `pipeline/common` is reserved for true cross-stage primitives (errors, contract helpers, artifact helpers), not domain-specific operations.
   - **Rationale:** Prevents layering violations and keeps imports aligned with architecture.

4. **Migrate incrementally with behavior-parity checks**
   - Move files in small batches, update imports/tests, run targeted and full suite checks after each batch.
   - Keep output contracts and failure semantics unchanged for parity-critical stages.
   - **Rationale:** Reduces refactor risk and keeps regressions observable.

## Risks / Trade-offs

- **[Risk] Migration churn introduces import/compile breakages** → **Mitigation:** Execute in small commits, run stage-focused tests plus full suite, and keep temporary aliases only where necessary.
- **[Risk] Misclassification of ownership (stage vs capability)** → **Mitigation:** Apply explicit placement rules and document representative examples for each rule.
- **[Risk] Partial migration leaves long-lived hybrid state** → **Mitigation:** Add completion criteria in tasks/specs that require `internal/scraper` cleanup and documentation updates before considering change done.
- **[Trade-off] Short-term diff size and review overhead** → **Mitigation:** Prefer mechanical moves + import updates first, followed by minimal structural cleanups.
