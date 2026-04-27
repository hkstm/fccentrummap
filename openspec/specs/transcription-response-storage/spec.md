## Purpose

Define canonical persistence behavior for Murmel transcription responses linked to stored article audio sources.

## Requirements

### Requirement: Store transcription responses linked to source audio
The repository SHALL persist every transcription attempt in `article_audio_transcriptions` and link it to `article_audio_sources.audio_source_id`.

#### Scenario: Insert transcription result
- **WHEN** a transcription request completes for a valid `audio_source_id`
- **THEN** the repository SHALL write a row containing provider, language, HTTP status, response payload, byte size, and timestamp
- **AND** the row SHALL reference the source audio via foreign key

### Requirement: Enforce uniqueness by source, provider, and language
The repository SHALL enforce uniqueness on `(audio_source_id, provider, language)` to prevent accidental duplicate rows.

#### Scenario: First result for tuple
- **WHEN** no row exists for `(audio_source_id, provider, language)`
- **THEN** the repository SHALL insert a new transcription row

#### Scenario: Repeat run for existing tuple
- **WHEN** a row already exists for `(audio_source_id, provider, language)`
- **THEN** the repository SHALL upsert that row instead of creating a duplicate
- **AND** it SHALL update payload and metadata to reflect the latest run

### Requirement: Store response payload as canonical JSON text
The repository SHALL store `response_json` as canonical JSON text suitable for JSON1 querying.

#### Scenario: Valid JSON response payload
- **WHEN** the repository receives a valid Murmel JSON response
- **THEN** it SHALL normalize/validate the payload as canonical JSON text before storage
- **AND** JSON1 functions (e.g. `json_extract`) SHALL operate on stored values without conversion errors

#### Scenario: Invalid or empty payload
- **WHEN** a transcription attempt returns an empty or non-JSON payload
- **THEN** the caller SHALL normalize the persisted payload to minimal valid JSON (for example `{}`)
- **AND** the repository SHALL persist the attempt with `http_status` and `error_message` for retry/inspection workflows
