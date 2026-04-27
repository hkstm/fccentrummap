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
go run ./cmd/scraper -db ../data/spots.db
go run ./cmd/export -db ../data/spots.db -out ../viz/public/data/spots.json

go run ./cmd/transcribe-audio --db-path ../data/spots.db --audio-source-id 1 --language nl
go run ./cmd/transcribe-audio --db-path ../data/spots.db --language nl

go run ./cmd/export-audio --db-path ../data/spots.db --audio-source-id 1 --out-dir ../data
go run ./cmd/export-transcription --db-path ../data/spots.db --transcription-id 1 --out-dir ../data

go run ./cmd/extract-spots-dry-run --db-path ../data/spots.db --transcription-id 1 --out-dir ../data
go run ./cmd/extract-spots-dry-run --db-path ../data/spots.db --use-latest --out-dir ../data

go test ./...
go build ./...
```

## Environment variables

- `MURMEL_API_KEY` (required by `cmd/transcribe-audio`; sent as `X-API-Key`)

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

If this document diverges from OpenSpec, treat `openspec/specs/` as the source of truth.
