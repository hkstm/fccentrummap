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

## Audio transcription CLI helpers

```bash
# Required for transcription requests
export MURMEL_API_KEY="..."

# Transcribe explicit audio source row
cd scraper && go run ./cmd/transcribe-audio --db-path ../data/spots.db --audio-source-id 1 --language nl

# Transcribe latest available audio source row
cd scraper && go run ./cmd/transcribe-audio --db-path ../data/spots.db --language nl

# Export source audio by id to data/
cd scraper && go run ./cmd/export-audio --db-path ../data/spots.db --audio-source-id 1 --out-dir ../data

# Export stored transcription JSON by id to data/
cd scraper && go run ./cmd/export-transcription --db-path ../data/spots.db --transcription-id 1 --out-dir ../data

# Dry-run extraction for an explicit transcription row
cd scraper && go run ./cmd/extract-spots-dry-run --db-path ../data/spots.db --transcription-id 1 --out-dir ../data

# Dry-run extraction for the latest transcription row (must be explicit)
cd scraper && go run ./cmd/extract-spots-dry-run --db-path ../data/spots.db --use-latest --out-dir ../data

# Re-run article text extraction for latest 5 articles and persist results
cd scraper && go run ./cmd/reextract-article-text --db-path ../data/spots.db --limit 5

# Preview extraction outcomes without writing to DB
cd scraper && go run ./cmd/reextract-article-text --db-path ../data/spots.db --limit 5 --dry-run

# Print extracted text segments for visual verification
cd scraper && go run ./cmd/reextract-article-text --db-path ../data/spots.db --limit 5 --dry-run --print-content

# Write extracted text report to file for review
cd scraper && go run ./cmd/reextract-article-text --db-path ../data/spots.db --limit 5 --dry-run --out-file ../data/article-text-report.txt
```

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
