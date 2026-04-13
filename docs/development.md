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
go test ./...
go build ./...
```

## Notes

- `data/spots.db` is local/generated data
- `viz/public/data/spots.json` is a generated artifact
- the frontend boundary is static JSON, not direct SQLite access

If this document diverges from OpenSpec, treat `openspec/specs/` as the source of truth.
