# Architecture

This document summarizes the implemented architecture.

## Canonical source

For normative behavior and architectural rules, use:

- `openspec/specs/project-layout/spec.md`
- `openspec/specs/web-scraper/spec.md`
- `openspec/specs/sqlite-storage/spec.md`
- `openspec/specs/static-data/spec.md`
- `openspec/specs/frontend-portability/spec.md`

## High-level pipeline

```text
scraper -> data/spots.db -> viz/public/data/spots.json -> viz frontend
```

## Current implementation shape

- `scraper/cmd/scrape` provides unified stage subcommands (`init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, `export-data`)
- Pipeline internals use explicit layering: **CLI → stage service → adapter**
- Stage-first packages live under `scraper/internal/pipeline/<stage>` and own stage DTOs, ports, service logic, and adapters
- Reusable capabilities live in explicit packages (for example `internal/audio`, `internal/contentfetch`, `internal/articletext`, `internal/transcription`, `internal/geocoder`, `internal/extraction`)
- Shared cross-stage helpers live under `scraper/internal/pipeline/common` and must stay domain-agnostic
- `scraper/internal/scraper` is deprecated for new business logic
- `export-data` reads SQLite and writes frontend JSON
- `viz/` is reserved for frontend work that consumes `/data/spots.json`

See `docs/service-package-organization.md` for the canonical placement rules and migration map.

## Current scraper processing flow (implemented)

Use `cmd/scrape` stage subcommands for orchestration. Stages are run explicitly in order, with SQLite as default mode and explicit file-mode handoff where configured.

Adapter behavior notes:
- SQLite adapter path remains the integrity anchor (FK/unique/transaction guarantees come from SQLite schema + repository behavior).
- File adapter path is geared toward deterministic handoff and ad-hoc debugging (human-inspectable JSON artifacts), not DB-equivalent integrity semantics.
- `geocode-spots` can now run in sqlite mode (persists geocodes + article spot links) or file mode (artifact-driven debugging/handoffs).

If this summary ever disagrees with `openspec/specs/`, treat the specs as canonical.
