## Why

The current pipeline is split across many CLIs with inconsistent entrypoints and mixed persistence behavior, which makes stage-by-stage operation and future extension harder than it needs to be. We need a unified command surface now so new functionality can be added into a stable workflow without repeatedly redesigning orchestration and I/O boundaries.

## What Changes

- Introduce a single top-level scraper CLI entrypoint with stage subcommands (`init`, `collect-article-urls`, `fetch-articles` with built-in text extraction, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, `export-data`).
- Define stage boundaries so each stage performs at most one expensive external I/O category (site crawl/fetch, YouTube acquisition, Murmel transcription, Gemma extraction, Google Places geocoding).
- Standardize default stage I/O to SQLite-backed inputs/outputs where implemented.
- Add optional file-based I/O mode per stage (explicit file input + deterministic output naming) for reproducibility, debugging, and offline handoff.
- For unsupported stage/mode combinations, fail fast with explicit non-zero errors (no silent no-ops).
- Define a stage-oriented artifact directory layout under `data/` that mirrors pipeline stages (with table references in filenames/metadata where useful).
- Keep this change scaffold-first: establish command and data-flow contracts even where implementation is currently partial or pending.
- Explicitly defer new SQLite geocode-to-final-`spots` writes in this change; use file-artifact handoff for geocode stage until follow-up integration.
- Add `init` preflight validation for required API environment variables to fail fast before expensive stage runs.

## Capabilities

### New Capabilities
- `unified-scrape-cli`: Single entrypoint and subcommand contract for stage-based scraping and enrichment workflows.
- `stage-artifact-file-io`: Optional file input/output mode per stage with standardized `data/` artifact layout and naming conventions.

### Modified Capabilities
- `web-scraper`: Align scraping stages with unified subcommands and SQLite/file mode execution paths.
- `youtube-audio-acquisition`: Align audio acquisition stage contract with unified CLI stage boundaries and I/O modes.
- `audio-transcription-cli`: Align transcription invocation and persistence behavior with unified stage execution.
- `transcript-spot-extraction`: Align extraction stage behavior with unified CLI orchestration and artifact/file I/O support.
- `google-places-text-search-geocoding`: Align geocoding stage invocation with unified command surface and stage contracts.
- `static-data`: Align export stage naming and behavior with unified CLI output workflow.
- `sqlite-storage`: Define stage-level read/write expectations and schema touchpoints for unified orchestration.

## Impact

- Affected code: `scraper/cmd/*` command structure, shared CLI plumbing, and stage orchestration code.
- Affected data flow: explicit contracts between stage outputs and inputs in both SQLite and file artifact modes.
- Affected docs/specs: multiple scraper-related capability specs plus new capability specs for unified CLI and artifact file I/O.
- Dependencies/systems: no new mandatory external providers, but clearer integration boundaries for fccentrum.nl, yt-dlp/YouTube, Murmel, Gemma, and Google Maps APIs.
