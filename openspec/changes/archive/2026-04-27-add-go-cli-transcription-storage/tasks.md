## 1. Schema and repository updates

- [x] 1.1 Extend SQLite initialization to create `article_audio_transcriptions` with foreign key to `article_audio_sources` and `UNIQUE(audio_source_id, provider, language)`
- [x] 1.2 Add JSON validity/canonicalization constraints for `response_json` storage and ensure startup creation is idempotent on existing DBs
- [x] 1.3 Implement repository query to fetch audio source by explicit `audio_source_id` including blob and format metadata
- [x] 1.4 Implement repository query to fetch latest audio source row with non-empty `audio_blob`
- [x] 1.5 Implement repository upsert for transcription rows keyed by `(audio_source_id, provider, language)` that updates payload and metadata

## 2. Murmel client integration

- [x] 2.1 Implement Murmel API client function that sends multipart form (`audio`, `language`) with `X-API-Key`
- [x] 2.2 Add startup validation for `MURMEL_API_KEY` with actionable error messages
- [x] 2.3 Capture HTTP status, response bytes, and transport/HTTP error details for persistence
- [x] 2.4 Normalize/validate response payload to canonical JSON text before repository write

## 3. CLI command implementation

- [x] 3.1 Add `transcribe-audio` command with flags `--audio-source-id`, `--language` (default `nl`), and optional `--db-path`
- [x] 3.2 Implement selection flow: explicit ID when provided, otherwise latest audio source
- [x] 3.3 Wire command execution to Murmel client and transcription upsert path
- [x] 3.4 Print clear command output for selected source ID, Murmel status, and saved transcription ID/result
- [x] 3.5 Ensure non-zero exits for missing IDs, missing API key, invalid JSON payload, and API failures

## 4. Export commands for local inspection

- [x] 4.1 Add command to export audio by `--audio-source-id` to `data/` with deterministic filename and extension
- [x] 4.2 Add command to export transcription JSON by `--transcription-id` to `data/` with deterministic filename
- [x] 4.3 Ensure exported transcription payload is UTF-8 JSON text and exported audio keeps exact stored bytes
- [x] 4.4 Add helpful CLI errors for unknown IDs and file write failures

## 5. Validation and documentation

- [x] 5.1 Validate schema creation and upsert behavior against existing `data/spots.db`
- [x] 5.2 Run end-to-end test: transcribe explicit ID and default-latest path, verify one row per `(audio_source_id, provider, language)`
- [x] 5.3 Run export command checks for both audio and transcription outputs in `data/`
- [x] 5.4 Document command usage examples and required environment variables in project docs/README
