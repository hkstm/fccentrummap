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
- `export-data` reads SQLite and writes frontend JSON
- `viz/` is reserved for frontend work that consumes `/data/spots.json`

## Current scraper processing flow (implemented)

Use `cmd/scrape` stage subcommands for orchestration. Stages are run explicitly in order, with SQLite as default mode and explicit file-mode handoff where configured.
If this summary ever disagrees with `openspec/specs/`, treat the specs as canonical.
