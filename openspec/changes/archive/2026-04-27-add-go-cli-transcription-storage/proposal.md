## Why

We can manually extract audio from SQLite and call Murmel, but this is error-prone and slow for repeat testing. We need a small Go CLI command that can run this flow end-to-end and persist transcription responses for reproducible local workflows.

## What Changes

- Add a Go CLI command to transcribe stored audio from `data/spots.db` via the Murmel API.
- Support selecting an `article_audio_sources.audio_source_id` explicitly, or defaulting to the latest available audio row.
- Add persistence for Murmel transcription responses as JSON blobs in a new SQLite table linked to the source audio row.
- Store request/response metadata (e.g., language, status, timestamps, byte size) to support inspection and retries.
- Keep existing audio acquisition and storage behavior unchanged (no breaking changes).

## Capabilities

### New Capabilities
- `audio-transcription-cli`: Provide a Go CLI workflow that selects an audio source row, submits audio to Murmel transcription, and reports success/failure to the user.
- `transcription-response-storage`: Persist Murmel JSON responses as BLOB/TEXT payloads with metadata and foreign-key linkage to `article_audio_sources`.

### Modified Capabilities
- `sqlite-storage`: Extend schema initialization and repository behavior to create and write/read a transcription-results table referencing `article_audio_sources`.

## Impact

- Affected code: Go CLI entrypoints/commands, SQLite repository layer, schema initialization/migrations, and DB write paths.
- APIs: Outbound POST to Murmel transcription API using API key from environment.
- Data model: New transcription table with foreign key to `article_audio_sources` and stored JSON payload.
- Dependencies/systems: Existing `modernc.org/sqlite` DB layer; Murmel API availability and auth configuration.