## 1. Gemini-direct extraction contract

- [x] 1.1 Define Gemini-direct prompt builder that uses only article URL and YouTube URL as primary inputs
- [x] 1.2 Define structured Gemini-direct response config/schema for presenter, place, YouTube timestamp seconds, evidence, confidence, and notes
- [x] 1.3 Implement parser/validator for Gemini-direct raw responses into a stable parsed candidate JSON shape
- [x] 1.4 Add unit tests for prompt contents, structured config, successful parsing, and malformed response diagnostics

## 2. Gemini-direct extraction stage

- [x] 2.1 Add stage DTOs, ports, and service package for `extract-spots-gemini-direct`
- [x] 2.2 Implement SQLite input adapter that selects article URL and associated YouTube URL without requiring cleaned article text or transcription rows
- [x] 2.3 Implement artifact writer for prompt, raw response, and parsed JSON outputs under a deterministic output directory
- [x] 2.4 Ensure Gemini-direct extraction does not write to canonical spot mention, geocode, correction, or export tables
- [x] 2.5 Add tests covering missing URL handling, artifact writes, and no canonical spot-data mutation

## 3. Comparison report stage

- [x] 3.1 Add stage DTOs, ports, and service package for `compare-spot-extractions`
- [x] 3.2 Implement baseline loader for transcript-based presenter and spot data from SQLite
- [x] 3.3 Implement Gemini-direct artifact loader with article-level malformed artifact notes
- [x] 3.4 Implement report assembly that lists baseline and Gemini-direct spots without automatic matching
- [x] 3.5 Implement Markdown renderer with summary counts and per-article presenter plus separate spot tables
- [x] 3.6 Add tests proving the report does not invoke Gemini or Google Places and does not emit matcher classifications

## 4. Unified CLI integration

- [x] 4.1 Add `extract-spots-gemini-direct` command with `--io sqlite`, `--db-path`, `--out-dir`, and `--model` flags
- [x] 4.2 Add `compare-spot-extractions` command with `--db-path`, `--gemini-direct-dir`, and `--out` flags
- [x] 4.3 Update stage mode validation so the new experimental commands are explicit and existing commands keep current defaults
- [x] 4.4 Add CLI tests for command availability, flag normalization, and unsupported mode errors

## 5. Documentation and verification

- [x] 5.1 Update README or supporting docs with the experimental workflow and example commands
- [x] 5.2 Run Go tests for scraper packages and fix regressions
- [x] 5.3 Run an end-to-end smoke test on a small article set when credentials/data are available and inspect generated artifacts/report
