## 1. Unified CLI scaffold

- [x] 1.1 Create `scraper/cmd/scrape` root command with subcommands: `init`, `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, `extract-spots`, `geocode-spots`, `export-data`
- [x] 1.2 Add shared flags and mode parsing (`--io`, `--db-path`, stage-specific `--in`) with default `--io sqlite`
- [x] 1.3 Implement central stage/mode validation so unsupported combinations fail non-zero with actionable errors

## 2. Init preflight and schema lifecycle

- [x] 2.1 Implement `scrape init` schema initialization/reset flow
- [x] 2.2 Add init-time required env validation (`MURMEL_API_KEY`, `GOOGLE_MAPS_API_KEY`, and one of `GEMINI_API_KEY`/`GOOGLE_API_KEY`/`GOOGLE_GENERATIVE_LANGUAGE_API_KEY`)
- [x] 2.3 Add clear missing-env diagnostics and tests for preflight failures

## 3. SQLite-first stage adapters

- [x] 3.1 Wire `collect-article-urls` sqlite mode to seed `articles_raw` (including single-article input path)
- [x] 3.2 Wire `fetch-articles` sqlite mode to fetch HTML, persist `video_id`, and persist article text extraction outcome/content
- [x] 3.3 Wire `acquire-audio` sqlite mode to existing audio acquisition + persistence behavior
- [x] 3.4 Wire `transcribe-audio` sqlite mode to existing Murmel transcription + persistence behavior
- [x] 3.5 Wire `extract-spots` sqlite mode to existing extraction record persistence behavior
- [x] 3.6 Wire `export-data` sqlite mode to existing export behavior

## 4. File-mode contracts and deterministic artifacts

- [x] 4.1 Implement deterministic identity-based artifact naming helper shared across file-mode stages
- [x] 4.2 Implement file-mode read/write adapters for `collect-article-urls`, `fetch-articles`, `acquire-audio`, `transcribe-audio`, and `extract-spots`
- [x] 4.3 Implement `geocode-spots` as file-mode supported stage with explicit `--in` requirement
- [x] 4.4 Ensure no implicit latest-artifact discovery in file mode; missing explicit `--in` fails validation

## 5. Geocode scope guardrails

- [x] 5.1 Enforce `geocode-spots --io sqlite` unsupported error path with explicit guidance to file mode
- [x] 5.2 Keep geocode SQLite writes to final `spots` deferred in this change (no new persistence integration)
- [x] 5.3 Emit deterministic geocode output artifacts for downstream/manual handoff

## 6. E2E validation, docs, and legacy removal

- [x] 6.1 Add/refresh CLI help text and README/development docs for unified command usage and mode support matrix
- [x] 6.2 Run single-article E2E validation flow using new commands; treat `export-data` as smoke test with potentially null payload
- [x] 6.3 Enforce no auto-retry behavior for expensive API stages in failure-handling flow
- [x] 6.4 Remove legacy command entrypoints/binaries and legacy-only code paths after parity validation
- [x] 6.5 Add/adjust tests for command routing, mode validation, env preflight, and stage smoke behaviors
