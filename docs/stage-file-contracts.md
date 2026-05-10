# Stage file-mode contracts

This document summarizes the typed file-mode contracts used by the stage adapters.

## Identity and artifact naming

- Identity is derived from explicit request identity when provided.
- Otherwise identity is derived from input filename prefix before `__`.
- Artifacts are written with deterministic stage/type suffixes via `cliutil.StageArtifactPath`.
- Payloads are intentionally human-inspectable JSON and do not include schema-version metadata.

## Stage contracts

- `collect-article-urls`
  - input: `{ "articleUrl": "..." }` or `{ "urls": ["..."] }`
  - output: `{ "identity": "...", "stage": "collectarticleurls", "articleUrls": [...] }`
- `fetch-articles`
  - input: `{ "articleUrls": ["..."] }`
  - output: `{ "identity": "...", "stage": "fetcharticles", "articleUrls": [...], "fetchedCount": n }`
- `acquire-audio`
  - input: `{ "articles": [{"url":"...","videoId":"..."}] }`
  - output: `{ "identity": "...", "stage": "acquireaudio", "acquired": [...] }`
- `transcribe-audio`
  - input: `{ "audioSourceId": n, "audioBlobBase64": "...", "language": "nl" }`
  - output: `{ "identity": "...", "stage": "transcribeaudio", "provider": "murmel", ... }`
- `extract-spots`
  - input: `{ "transcriptionJson": "...", "articleText": "..." }`
  - output: `{ "identity": "...", "stage": "extractspots", "presenterName": ..., "spots": [...] }`
- `geocode-spots`
  - input: `{ "query": "..." }`
  - output: `{ "identity": "...", "stage": "geocodespots", "query": "...", "coordinates": {...} }`
  - note: file-mode geocoding requires `GOOGLE_MAPS_API_KEY`
- `export-data`
  - input: `{ "authors": [...], "spots": [...] }`
  - output: `{ "identity": "...", "stage": "exportdata", "authors": [...], "spots": [...] }`

## Backend differences

- SQLite mode is the canonical integrity backend.
- File mode is for deterministic handoff/debugging and cannot fully replicate DB constraints.
- `geocode-spots --io sqlite` is intentionally unsupported in this scope.
- `geocode-spots --io file` still requires `GOOGLE_MAPS_API_KEY`.
