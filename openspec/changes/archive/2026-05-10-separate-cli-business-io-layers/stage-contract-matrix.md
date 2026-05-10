# Stage contract matrix (implementation guide)

## 1) collectarticleurls
- Request:
  - `dbPath` (sqlite)
  - `articleURL` (optional seed)
  - `inputPath` (file; optional when `articleURL` provided)
  - `identity` (optional override)
- File input contract:
  - `{ "articleUrl": "..." }` or `{ "urls": ["..."] }`
- File output contract:
  - `{ "identity": "...", "stage": "collectarticleurls", "articleUrls": ["..."] }`

## 2) fetcharticles
- Request: `dbPath` (sqlite), `inputPath` (file), `identity` (optional)
- File input contract:
  - `{ "articleUrls": ["..."] }`
- File output contract:
  - `{ "identity": "...", "stage": "fetcharticles", "articleUrls": ["..."], "fetchedCount": n }`

## 3) acquireaudio
- Request: `dbPath` (sqlite), `inputPath` (file), `identity` (optional)
- File input contract:
  - `{ "articleUrls": ["..."] }` or `{ "articles": [{"url":"...","videoId":"..."}] }`
- File output contract:
  - `{ "identity": "...", "stage": "acquireaudio", "acquired": [{"url":"...","videoId":"..."}] }`

## 4) transcribeaudio
- Request: `dbPath` (sqlite), `language` (sqlite), `inputPath` (file), `identity` (optional)
- File input contract:
  - `{ "audioSourceId": n, "language": "nl", "audioFormat": "wav", "audioBlobBase64": "..." }`
- File output contract:
  - `{ "identity": "...", "stage": "transcribeaudio", "provider":"murmel", "httpStatus": n, "responseJson": "...", "errorMessage": "..." }`

## 5) extractspots
- Request: `dbPath` (sqlite), `outDir` / `gemmaModel` / `apiKey` / `endpoint` (sqlite), `inputPath` (file), `identity` (optional)
- File input contract:
  - `{ "transcriptionJson": "...", "articleText": "..." }`
- File output contract:
  - `{ "identity": "...", "stage": "extractspots", "presenterName": "...", "spots": [...] }`

## 6) geocodespots
- Request: `inputPath` (file), `identity` (optional)
- SQLite mode: unsupported by design (must return actionable error)
- File input contract:
  - `{ "query": "..." }`
- File output contract:
  - `{ "identity": "...", "stage": "geocodespots", "query": "...", "coordinates": {...} }`

## 7) exportdata
- Request: `dbPath` (sqlite), `outputPath` (sqlite), `inputPath` (file), `identity` (optional)
- File input contract:
  - `{ "authors": [...], "spots": [...] }`
- File output contract:
  - `{ "identity": "...", "stage": "exportdata", "authors": [...], "spots": [...] }`

Notes:
- Identity in file mode derives from `request.identity` first, else from input filename prefix before `__`.
- Output artifact names remain deterministic via `cliutil.StageArtifactPath`.
- Contracts intentionally remain human-inspectable and avoid schema-version metadata.
