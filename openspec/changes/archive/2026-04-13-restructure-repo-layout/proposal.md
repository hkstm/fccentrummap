## Why

The repository is currently organized around the initial Go scraper, but the planned system now has two durable concerns: a Go data pipeline and a future frontend that consumes generated static data. The current top-level layout makes that boundary unclear and leaves too much repo-level workflow implicit.

## What Changes

- Reorganize the repository around top-level `scraper/`, `viz/`, `docs/`, and `data/` areas
- Move the Go module root into `scraper/`
- Standardize generated artifact locations so the SQLite database lives under `data/` and the exported frontend data lives at `viz/public/data/spots.json`
- Add a root `Makefile` as the canonical entrypoint for repo-level workflows
- Establish frontend portability constraints so future Next.js work remains host-portable and consumes generated JSON rather than SQLite directly

## Capabilities

### New Capabilities
- `project-layout`: A top-level repository structure that separates the Go pipeline, frontend area, generated data, and durable documentation
- `developer-workflow`: Canonical repo-level commands for scrape, export, build, and verification
- `frontend-portability`: Constraints for future frontend work so Next.js is used as a framework, not as a Vercel-specific platform

### Modified Capabilities
- `static-data`: Update exporter and artifact path assumptions to match the rewritten repository layout

## Impact

- **Repository structure**: Go code moves under `scraper/`; frontend work belongs under `viz/`
- **Artifact locations**: SQLite output moves under `data/`; generated JSON moves to `viz/public/data/spots.json`
- **Developer workflow**: Root `Makefile` becomes the main entrypoint for common tasks
- **Frontend boundary**: Frontend work continues to consume exported static JSON and does not talk to SQLite directly
