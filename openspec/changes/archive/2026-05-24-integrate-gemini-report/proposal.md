## Why

Gemini appears capable of answering article/video spot questions directly from article and YouTube context, which may let us simplify the expensive transcription-driven extraction pipeline. We need a safe way to evaluate that opportunity by running a Gemini-direct extraction path alongside the current pipeline and producing a comparison report before replacing any existing behavior.

## What Changes

- Add an experimental Gemini-direct spot extraction path that uses scraped article context plus the article's YouTube URL as model input, without requiring Murmel transcription output.
- Persist Gemini-direct prompts, raw responses, and parsed candidate results as inspectable artifacts so runs are reproducible and reviewable.
- Add a comparison report that evaluates Gemini-direct results against the existing transcript-based extraction for the same articles.
- Compare presenter extraction, place overlap, geocoded identity overlap where available, timestamp differences, and candidate-only/baseline-only spots.
- Keep the existing transcript-based extraction, geocoding, corrections, and JSON export behavior unchanged during the experiment.
- Do not make Gemini-direct results the default source for `spots.json` in this change.

## Capabilities

### New Capabilities
- `gemini-direct-spot-extraction`: Experimental extraction of presenter and timestamped spot candidates from article and YouTube context using Gemini directly.
- `spot-extraction-comparison-report`: Generation of reviewable reports comparing Gemini-direct extraction output with the existing transcript-based extraction output.

### Modified Capabilities
- `unified-scrape-cli`: Expose the experimental Gemini-direct extraction and comparison report operations through explicit CLI stage commands without changing existing stage defaults.

## Impact

- Affected Go code under `scraper/cmd/scrape`, `scraper/internal/genai`, extraction-related packages, and likely new pipeline stage packages.
- New artifact files under `data/` or a deterministic subdirectory for Gemini-direct prompts, raw responses, parsed outputs, and comparison reports.
- Existing SQLite schema may need additive storage for experimental extraction runs, or the first implementation may keep Gemini-direct outputs artifact-only if sufficient for comparison.
- Requires Gemini API credentials already supported by the project.
- No breaking change to existing `extract-spots`, `geocode-spots`, `export-data`, or frontend `spots.json` consumption.
