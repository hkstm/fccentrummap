## 1. Repository layout

- [x] 1.1 Move the Go module root and source tree under `scraper/`
- [x] 1.2 Create durable top-level homes for `viz/`, `docs/`, and `data/`
- [x] 1.3 Align ignored/generated artifact paths with the new layout

## 2. Workflow entrypoints

- [x] 2.1 Add a root `Makefile` with canonical repo-level commands for scrape, export, build, and checks
- [x] 2.2 Update contributor-facing documentation to describe the new layout and command surface

## 3. Frontend boundary and portability

- [x] 3.1 Document `viz/` as the frontend area that consumes generated JSON rather than SQLite
- [x] 3.2 Add portable Next.js guidance that avoids Vercel-only dependencies and preserves generic hosting options

## 4. OpenSpec alignment

- [x] 4.1 Update affected `static-data` path assumptions to match the new layout
- [x] 4.2 Verify related OpenSpec changes still describe the Go → JSON → frontend boundary correctly
