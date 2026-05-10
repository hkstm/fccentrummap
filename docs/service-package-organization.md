# Service package organization

## Canonical ownership model

- **Stage-owned logic**: `scraper/internal/pipeline/<stage>`
  - Owns stage DTOs, orchestration service, and stage ports/adapters.
- **Capability-owned reusable logic**: explicit capability packages under `scraper/internal/*`
  - Current canonical capabilities:
    - `scraper/internal/audio`: audio download/acquisition business logic
    - `scraper/internal/contentfetch`: article URL crawling + article fetch/persist flow
    - `scraper/internal/articletext`: article text extraction logic
    - `scraper/internal/transcription`: transcription capability
    - `scraper/internal/geocoder`: geocoding capability
    - `scraper/internal/extraction`: spot extraction capability
- **Cross-stage primitives only**: `scraper/internal/pipeline/common`
  - Reserved for generic contracts, errors, and artifact IO helpers.

## Legacy package deprecation

`scraper/internal/scraper` is deprecated as a destination for new business logic.

Rules:
- Do not add new service/business logic to `scraper/internal/scraper`.
- Place new logic in a stage package or explicit capability package.
- Stage packages must not depend directly on `scraper/internal/scraper`.

## Inventory and target classification (2026-05-10)

### Stage-owned (`scraper/internal/pipeline/<stage>`)

- `acquireaudio/*`
- `collectarticleurls/*`
- `fetcharticles/*`
- `transcribeaudio/*`
- `extractspots/*`
- `geocodespots/*`
- `exportdata/*`

### Cross-stage primitives (`scraper/internal/pipeline/common`)

- `artifactio.go`
- `contracts.go`
- `errors.go`

### Capability-owned reusable logic

- `scraper/internal/audio/audio.go`
- `scraper/internal/contentfetch/contentfetch.go`
- `scraper/internal/contentfetch/youtube.go`
- `scraper/internal/articletext/extractor.go`
- `scraper/internal/extraction/*`
- `scraper/internal/transcription/*`
- `scraper/internal/geocoder/*`

### Deprecated legacy location

- `scraper/internal/scraper/*` (retained only during migration cleanup; no new logic)

## Placement checklist for contributors

Before adding or moving code:

1. Is it stage orchestration/DTO/ports for a single stage?
   - Put it in `scraper/internal/pipeline/<stage>`.
2. Is it reusable domain logic across stages?
   - Put it in an explicit capability package.
3. Is it truly generic cross-stage plumbing?
   - Put it in `scraper/internal/pipeline/common`.
4. Does this add or keep imports to `scraper/internal/scraper` from stage packages?
   - If yes, stop and migrate to capability boundaries first.
