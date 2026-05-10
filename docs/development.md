# Development workflow

## Canonical source

The authoritative workflow expectations live in:

- `openspec/specs/developer-workflow/spec.md`

## Common commands

```bash
make scrape
make export
make build
make check
```

## Direct Go commands

```bash
cd scraper

# Unified stage CLI (urfave/cli v3)
go run ./cmd/scrape --help
go run ./cmd/scrape <stage> --help

# Required env for init preflight: MURMEL_API_KEY, GOOGLE_MAPS_API_KEY,
# and one of GEMINI_API_KEY / GOOGLE_API_KEY / GOOGLE_GENERATIVE_LANGUAGE_API_KEY
go run ./cmd/scrape init --db-path ../data/spots.db --reset
go run ./cmd/scrape collect-article-urls --io sqlite --db-path ../data/spots.db --article-url "<FCCENTRUM_ARTICLE_URL>"
go run ./cmd/scrape fetch-articles --io sqlite --db-path ../data/spots.db
go run ./cmd/scrape acquire-audio --io sqlite --db-path ../data/spots.db
go run ./cmd/scrape transcribe-audio --io sqlite --db-path ../data/spots.db --language nl
go run ./cmd/scrape extract-spots --io sqlite --db-path ../data/spots.db --out-dir ../data

# Geocode stage is file-mode only for now
go run ./cmd/scrape geocode-spots --io file --in ../data/stages/extract-spots/<identity>__extract-spots__candidates.json

# Export smoke test
go run ./cmd/scrape export-data --io sqlite --db-path ../data/spots.db --out ../viz/public/data/spots.json

go test ./...
go build ./...
```

## Environment variables

- `MURMEL_API_KEY` (required by `cmd/scrape init` preflight and transcription)
- `GOOGLE_MAPS_API_KEY` (required by `cmd/scrape init` preflight and geocode)
- one of `GEMINI_API_KEY` / `GOOGLE_API_KEY` / `GOOGLE_GENERATIVE_LANGUAGE_API_KEY` (required by `cmd/scrape init` preflight)
- `GOOGLE_PLACES_TEXT_SEARCH_ENDPOINT` (optional override for geocoder endpoint; useful for local testing)

## Commit message convention

This repo uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

- Format: `<type>[optional scope][!]: <description>`
- Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`

Examples:

```text
feat(scraper): ingest youtube audio blobs
fix(repository): keep audio inserts idempotent
docs: describe git hook setup
```

Basic validation is enforced via the local `commit-msg` hook in `.githooks/commit-msg`.
Set it up once per clone:

```bash
make setup-hooks
```

## Notes

- `data/spots.db` is local/generated data
- `viz/public/data/spots.json` is a generated artifact
- the frontend boundary is static JSON, not direct SQLite access
- `scrape` enforces stage/mode validation before processing and fails non-zero with actionable guidance for unsupported combinations
- `scrape geocode-spots --io sqlite` is intentionally unsupported in this scope; use `--io file --in <path>`
- legacy stdlib `flag` wiring compatibility shims are intentionally removed; prefer documented urfave/cli v3 invocation forms
- geocoder requests enforce a hard `locationRestriction.rectangle` (low `52.274525,4.711585`; high `52.461764,5.073559`)

If this document diverges from OpenSpec, treat `openspec/specs/` as the source of truth.
