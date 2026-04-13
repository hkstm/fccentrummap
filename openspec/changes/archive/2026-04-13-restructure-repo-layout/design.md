## Context

The existing repository started as a single Go scraper project rooted at the repo top level. The planned system now spans a scraper pipeline, generated artifacts, and a future visualization app. Existing OpenSpec changes already assume a build-time boundary between Go and the frontend, but some current path assumptions still reflect the old layout.

This change captures the repository-level structure needed to make those concerns explicit without changing the core product behavior.

## Goals / Non-Goals

**Goals:**
- Make `scraper/` the Go module root
- Reserve `viz/` as the frontend application area
- Normalize generated artifact locations under durable top-level conventions
- Provide one obvious repo-level command surface
- Keep the frontend contract as static JSON, not direct database access
- Preserve frontend host portability

**Non-Goals:**
- Building the full frontend application
- Changing the JSON data contract beyond path/layout updates
- Introducing runtime coupling between the frontend and SQLite
- Replacing OpenSpec changes that define scraping, export, or map behavior

## Decisions

### 1. `scraper/` becomes the Go module root
**Choice:** Move the current Go module and source tree under `scraper/`.

**Rationale:** The Go pipeline is one subsystem of the repo, not the repo itself. Making `scraper/` the module root leaves room for `viz/` and other top-level concerns without mixing implementation boundaries.

### 2. Generated artifacts get stable, role-based locations
**Choice:** Store the scraper database under `data/` and the frontend export under `viz/public/data/spots.json`.

**Rationale:** Generated artifacts should live in predictable locations that match their consumers. `data/` is the durable home for local/generated pipeline outputs, while `viz/public/data/spots.json` is the explicit handoff artifact for the frontend.

### 3. The root `Makefile` is the canonical workflow surface
**Choice:** Define repo-level commands at the repository root and delegate into `scraper/` and `viz/` as needed.

**Rationale:** Contributors should not need to memorize subsystem-specific working directories for common workflows. A root entrypoint also keeps docs short and consistent.

### 4. The frontend boundary remains static JSON
**Choice:** Future frontend work in `viz/` consumes generated JSON and does not read SQLite directly.

**Rationale:** This preserves the architecture already established by `render-spots-map`: Go handles SQLite access and exports static data; the frontend remains portable and deployable without a database runtime.

### 5. Next.js is allowed as a framework, not as a platform dependency
**Choice:** Capture portability constraints for the future `viz/` area: avoid Vercel-only services and keep business logic in plain TypeScript modules.

**Rationale:** The repo should support local development, generic Docker/container deployment, and non-Vercel static or Node hosting options.

## Risks / Trade-offs

- **Broad file movement**: Repository-wide relocation can temporarily break commands and imports. Mitigation: define the new module root and command entrypoints explicitly.
- **Stale path references**: Existing docs/spec text may still point at top-level Go paths. Mitigation: update affected OpenSpec requirements in this change.
- **Frontend over-coupling risk**: A future app may drift toward direct DB access or host-specific services. Mitigation: codify the JSON boundary and portability constraints here.
