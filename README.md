# fccentrummap

FC Centrum map data pipeline and future visualization app.

## Single source of truth

Canonical project behavior and architectural constraints live in:

- `openspec/specs/`

Use other Markdown files for explanation and workflow help, not as competing sources of truth.

## Repo layout

```text
.
├── data/        # local/generated pipeline artifacts
├── docs/        # supporting documentation
├── openspec/    # canonical specs and change history
├── scraper/     # Go module for scraping and export
└── viz/         # future frontend area
```

## Canonical commands

```bash
make scrape
make export
make build
make check
```

## Key spec areas

- `openspec/specs/project-layout/spec.md`
- `openspec/specs/developer-workflow/spec.md`
- `openspec/specs/web-scraper/spec.md`
- `openspec/specs/sqlite-storage/spec.md`
- `openspec/specs/static-data/spec.md`
- `openspec/specs/frontend-portability/spec.md`

## Supporting docs

- `docs/architecture.md`
- `docs/development.md`
- `docs/frontend-portability.md`
- `viz/README.md`
