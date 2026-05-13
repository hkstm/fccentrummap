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
make setup-hooks
```

## Unified scrape CLI (stage-based)

The unified CLI is wired with `urfave/cli/v3` and uses subcommands plus idiomatic `--help` output.

```bash
cd scraper

go run ./cmd/scrape --help
go run ./cmd/scrape <stage> --help

# Preflight env + schema init (fails fast on missing required API keys)
go run ./cmd/scrape init --db-path ../data/spots.db --reset

# SQLite-first stages
go run ./cmd/scrape collect-article-urls --io sqlite --db-path ../data/spots.db --article-url "<FCCENTRUM_ARTICLE_URL>"
go run ./cmd/scrape fetch-articles --io sqlite --db-path ../data/spots.db
go run ./cmd/scrape acquire-audio --io sqlite --db-path ../data/spots.db
go run ./cmd/scrape transcribe-audio --io sqlite --db-path ../data/spots.db --language nl
go run ./cmd/scrape extract-spots --io sqlite --db-path ../data/spots.db --out-dir ../data

# Geocode is file-mode only in this change scope
# (sqlite mode fails with explicit guidance)
go run ./cmd/scrape geocode-spots --io file --in ../data/stages/extract-spots/<identity>__extract-spots__candidates.json

# Export smoke test (can succeed even with null payload during scaffold phase)
go run ./cmd/scrape export-data --io sqlite --db-path ../data/spots.db --out ../viz/public/data/spots.json
```

### Stage mode support matrix

| Stage | `--io sqlite` | `--io file` |
|---|---|---|
| `init` | Supported | Not supported (error) |
| `collect-article-urls` | Supported | Not implemented (error; use sqlite) |
| `fetch-articles` | Supported | Not implemented (error; use sqlite) |
| `acquire-audio` | Supported | Not implemented (error; use sqlite) |
| `transcribe-audio` | Supported | Not implemented (error; use sqlite) |
| `extract-spots` | Supported | Not implemented (error; use sqlite) |
| `geocode-spots` | Not supported (error) | Supported |
| `export-data` | Supported | Not implemented (error; use sqlite) |

## Notes

Legacy standalone command entrypoints were removed in favor of `cmd/scrape` stage subcommands.
Deprecated stdlib `flag` parsing paths and legacy invocation shims were intentionally dropped during the urfave/cli v3 migration.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

- `<type>[optional scope][!]: <description>`
- Example: `feat(scraper): add youtube audio acquisition stage`

Enable local commit-message validation hooks with:

```bash
make setup-hooks
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
