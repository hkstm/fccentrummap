## Purpose

Define CLI behavior for transcribing stored audio sources through Murmel and exporting stored audio/transcription artifacts for local inspection.

## Requirements

### Requirement: CLI transcribes selected stored audio through Murmel
The system SHALL provide a Go CLI command that transcribes audio from `article_audio_sources` by explicit ID or by defaulting to the latest available source.

#### Scenario: Transcribe by explicit audio source ID
- **WHEN** the user runs the transcription command with `--audio-source-id <id>` and the row exists
- **THEN** the CLI SHALL load that row's audio payload and metadata
- **AND** it SHALL send the payload to Murmel for transcription

#### Scenario: Transcribe latest audio when ID omitted
- **WHEN** the user runs the transcription command without `--audio-source-id`
- **THEN** the CLI SHALL select the latest `article_audio_sources` row with non-empty `audio_blob`
- **AND** it SHALL use that row as the transcription input

#### Scenario: Missing requested audio source
- **WHEN** the user provides `--audio-source-id <id>` that does not exist
- **THEN** the CLI SHALL exit with a non-zero status
- **AND** it SHALL print a clear error indicating the missing ID

### Requirement: CLI uses configured Murmel API contract
The system SHALL call Murmel using multipart upload with `audio` and `language` fields and authenticate using the `X-API-Key` header.

#### Scenario: Valid API key and successful request
- **WHEN** `MURMEL_API_KEY` is configured and Murmel returns HTTP 2xx
- **THEN** the CLI SHALL treat the request as successful
- **AND** it SHALL persist the transcription response record

#### Scenario: Missing API key
- **WHEN** `MURMEL_API_KEY` is missing or empty
- **THEN** the CLI SHALL fail before sending the request
- **AND** it SHALL print actionable guidance about setting the environment variable

### Requirement: CLI exports stored artifacts to the data directory
The system SHALL provide export commands for both source audio and transcription JSON artifacts by explicit IDs.

#### Scenario: Export audio by audio source ID
- **WHEN** the user requests audio export with `--audio-source-id <id>`
- **THEN** the CLI SHALL write the audio payload to `data/` with a deterministic filename containing that ID
- **AND** it SHALL preserve the correct file extension based on stored format/metadata

#### Scenario: Export transcription JSON by transcription ID
- **WHEN** the user requests transcription export with `--transcription-id <id>`
- **THEN** the CLI SHALL write the canonical JSON payload to `data/` with a deterministic filename containing that ID
- **AND** it SHALL produce valid UTF-8 JSON text
