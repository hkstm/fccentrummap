## Context

The current `scrape` command mixes three responsibilities in one place: CLI parsing/routing, stage orchestration, and persistence details. Some stage logic already lives in service-like packages (`internal/scraper`, `internal/extractspots`), but most flows are still coupled to concrete SQLite repository types, while file mode is often passthrough scaffolding rather than equivalent behavior.

This change introduces a layered architecture so stage behavior is defined once in business services and then executed through I/O adapters (SQLite or file artifacts). The goal is predictable stage semantics, clearer testability, and fewer regressions when adding new stages or changing persistence behavior.

Constraints:
- Keep existing stage names and user-facing stage sequence intact.
- Preserve SQLite integrity semantics (FK/constraints/uniqueness) as primary data guarantees.
- Allow file mode to implement the same stage contracts, with explicit limitations where strict DB guarantees are not possible.

## Goals / Non-Goals

**Goals:**
- Separate CLI concerns from business orchestration and persistence concerns.
- Define typed stage contracts for inputs/outputs independent of storage backend.
- Introduce adapter interfaces that can be implemented by SQLite-backed and file-backed adapters.
- Replace file-mode passthrough behavior with contract-driven stage execution.
- Improve testability via contract tests that run the same stage behavior against multiple adapters.

**Non-Goals:**
- Rewriting scraping/transcription/extraction/geocoding domain algorithms.
- Guaranteeing identical performance characteristics between SQLite and file adapters.
- Removing SQLite as the default/primary backend for production-like workflows.

## Decisions

1. **Adopt explicit three-layer boundary: CLI → Services → Adapters**
   - CLI layer (`cmd/scrape`) handles flags, mode selection, and user-facing errors only.
   - Service layer defines stage use-cases and domain-level input/output contracts.
   - Adapter layer implements persistence and handoff operations for SQLite and file artifacts.
   - **Alternative considered:** Keep current mixed orchestration with incremental cleanup. Rejected due to continued coupling and uneven file-mode behavior.

2. **Use backend-agnostic interfaces at service boundaries**
   - Define narrow interfaces per stage/use-case (e.g., article source loading, audio source persistence, transcription persistence, artifact load/save).
   - Services depend on interfaces, not `*repository.Repository`.
   - **Alternative considered:** one large repository interface for all stages. Rejected because it recreates tight coupling and increases mocking complexity.

3. **Treat file mode as first-class adapter with typed artifact contracts**
   - Define per-stage artifact schemas and deterministic naming/identity rules.
   - File adapter must parse/validate inputs and emit typed outputs rather than raw passthrough copies.
   - **Alternative considered:** keep generic passthrough in file mode. Rejected because it provides weak guarantees and unclear handoff semantics.

4. **Keep SQLite adapter as integrity anchor**
   - Existing SQLite constraints (FKs/checks/uniques) remain the canonical integrity model.
   - File adapter documents weaker guarantees and uses validation checks to reduce divergence.
   - **Alternative considered:** emulate transactional integrity in file mode. Rejected for complexity and limited value relative to project needs.

5. **Migrate stage-by-stage behind stable CLI commands**
   - Preserve command names and high-level invocation patterns while routing internals to new services/adapters.
   - Allow temporary adapter capability gaps per stage if explicitly documented and validated.
   - **Alternative considered:** big-bang rewrite of all stages at once. Rejected due to high regression risk.

## Proposed Directory Structure

```text
scraper/
├─ cmd/
│  └─ scrape/
│     └─ main.go                       # CLI wiring only: flags, command routing, mode selection
└─ internal/
   └─ pipeline/
      ├─ common/
      │  ├─ contracts.go               # Small shared interfaces/types used by multiple stages
      │  ├─ errors.go                  # Shared domain error types/mapping helpers
      │  └─ artifactio.go              # Generic artifact path/read/write helpers (non-stage-specific)
      │
      ├─ collectarticleurls/
      │  ├─ collectarticleurls_service.go                 # Stage business logic entrypoint
      │  ├─ collectarticleurls_dto.go                     # Stage input/output contracts
      │  ├─ collectarticleurls_ports.go                   # Stage-required interfaces (read/write operations)
      │  ├─ collectarticleurls_sqlite_adapter.go          # SQLite implementation of stage ports
      │  ├─ collectarticleurls_file_adapter.go            # File artifact implementation of stage ports
      │  └─ collectarticleurls_service_test.go            # Stage tests against mocked ports
      │
      ├─ fetcharticles/
      │  ├─ fetcharticles_service.go
      │  ├─ fetcharticles_dto.go
      │  ├─ fetcharticles_ports.go
      │  ├─ fetcharticles_sqlite_adapter.go
      │  └─ fetcharticles_file_adapter.go
      │
      ├─ acquireaudio/
      │  ├─ acquireaudio_service.go
      │  ├─ acquireaudio_dto.go
      │  ├─ acquireaudio_ports.go
      │  ├─ acquireaudio_sqlite_adapter.go
      │  └─ acquireaudio_file_adapter.go
      │
      ├─ transcribeaudio/
      │  ├─ transcribeaudio_service.go
      │  ├─ transcribeaudio_dto.go
      │  ├─ transcribeaudio_ports.go
      │  ├─ transcribeaudio_sqlite_adapter.go
      │  └─ transcribeaudio_file_adapter.go
      │
      ├─ extractspots/
      │  ├─ extractspots_service.go
      │  ├─ extractspots_dto.go
      │  ├─ extractspots_ports.go
      │  ├─ extractspots_sqlite_adapter.go
      │  └─ extractspots_file_adapter.go
      │
      ├─ geocodespots/
      │  ├─ geocodespots_service.go
      │  ├─ geocodespots_dto.go
      │  ├─ geocodespots_ports.go
      │  ├─ geocodespots_sqlite_adapter.go
      │  └─ geocodespots_file_adapter.go
      │
      ├─ exportdata/
      │  ├─ exportdata_service.go
      │  ├─ exportdata_dto.go
      │  ├─ exportdata_ports.go
      │  ├─ exportdata_sqlite_adapter.go
      │  └─ exportdata_file_adapter.go
      │
      └─ contracttests/
         ├─ stage_parity_test.go       # Same stage scenarios executed against sqlite + file adapters
         └─ fixtures/                  # Shared fixture payloads/artifacts for parity tests
```

Notes on component types:
- `<stage>_service.go`: pure stage orchestration/business flow; no direct DB/file dependency.
- `<stage>_dto.go`: stage-owned request/response structs for deterministic contracts.
- `<stage>_ports.go`: minimal interfaces the service needs (dependency inversion boundary).
- `<stage>_sqlite_adapter.go`: maps stage ports to `internal/repository` and SQLite semantics.
- `<stage>_file_adapter.go`: maps stage ports to file artifacts for ad-hoc debugging workflows.
- `contracttests/*`: validates backend parity for all 7 parity-critical stages.

## Risks / Trade-offs

- **[Risk] Interface design becomes too broad or leaky** → **Mitigation:** keep interfaces use-case-specific and review against concrete call sites.
- **[Risk] Behavior drift between SQLite and file adapters** → **Mitigation:** add contract tests for stage invariants and deterministic artifact outputs.
- **[Risk] Migration complexity across multiple modules** → **Mitigation:** migrate one stage at a time with compatibility checks and progressive tasking.
- **[Risk] More abstractions increase short-term code volume** → **Mitigation:** prioritize clarity, remove obsolete paths as each stage migration stabilizes.

## Migration Plan

1. Define service contracts and adapter interfaces for all unified stages.
2. Build SQLite adapter wrappers around existing repository methods first.
3. Implement file adapter contracts for current stage artifact flows (replace passthrough behavior).
4. Refactor CLI subcommands to call services with selected adapter based on `--io` mode.
5. Add contract tests per stage to validate backend-agnostic behavior.
6. Remove obsolete mixed-layer code paths once each stage is migrated.

Rollback strategy:
- Keep migration in incremental commits; revert per-stage refactor if behavior regressions occur.
- Preserve previous SQLite-oriented path until each stage’s new service route passes checks.

## Open Questions

- None currently.

## Confirmed Scope Decisions

- Milestone 1 parity-critical coverage includes all seven unified pipeline stages in file mode:
  - `collect-article-urls`
  - `fetch-articles`
  - `acquire-audio`
  - `transcribe-audio`
  - `extract-spots`
  - `geocode-spots`
  - `export-data`
- `init` remains intentionally SQLite-only.
- File artifacts remain primarily a debugging/handoff mechanism (ad-hoc usage), not a long-term versioned public interface.
- Do not add explicit schema-version metadata to artifact payloads in this milestone; keep artifact structure simple and optimize for practical debugging clarity over formal compatibility machinery.
- Organize pipeline code by stage-first package boundaries: each stage package owns its DTOs/contracts and stage service logic.
- Keep cross-stage shared code at the same layer in dedicated common packages, restricted to truly generic primitives/utilities used by multiple stages.
