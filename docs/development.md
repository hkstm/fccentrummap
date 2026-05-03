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

go run ./cmd/geocode-place --query "Dam Square Amsterdam"
# Example success JSON:
# {"query":"Dam Square Amsterdam","name":"Dam","placeId":"...","mapsUrl":"https://www.google.com/maps/search/?api=1&query=Dam+Square+Amsterdam&query_place_id=..."}

go test ./...
go build ./...
```

## Environment variables

- `MURMEL_API_KEY` (required by `cmd/transcribe-audio`; sent as `X-API-Key`)
- `GOOGLE_MAPS_API_KEY` (required by `cmd/geocode-place` and geocoder)
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
- `cmd/geocode-place` enforces a hard `locationRestriction.rectangle` in the text-search request (low `52.274525,4.711585`; high `52.461764,5.073559`); if nothing matches inside that box, it returns a deterministic no-result error in JSON
- `cmd/geocode-place` success output contains `query`, `name`, `placeId`, and a stable `mapsUrl` derived from query + place ID

If this document diverges from OpenSpec, treat `openspec/specs/` as the source of truth.
