## ADDED Requirements

### Requirement: Audio acquisition uses yt-dlp for embedded YouTube videos
The pipeline SHALL acquire audio for detected embedded YouTube videos using `yt-dlp`.

#### Scenario: Video ID available
- **WHEN** an article has a detected `video_id`
- **THEN** the pipeline SHALL invoke `yt-dlp` to download an audio artifact suitable for local ingestion

#### Scenario: Video ID missing
- **WHEN** an article has no detected `video_id`
- **THEN** the pipeline SHALL skip audio acquisition for that article without failing the entire run

### Requirement: Audio format policy supports practical downstream usage
The acquisition stage SHALL prefer WAV when available and allow fallback formats from a defined accepted set.

#### Scenario: Preferred format available
- **WHEN** the runtime environment supports extraction/conversion to WAV
- **THEN** the pipeline SHALL persist WAV metadata for that acquired artifact

#### Scenario: Preferred format unavailable
- **WHEN** WAV extraction is unavailable for an article/video
- **THEN** the pipeline SHALL accept and persist one of: `m4a`, `mp3`, `flac`, `ogg`, `webm`, `mp4`

### Requirement: Acquisition failures are explicit and retryable
Audio-acquisition failures SHALL be surfaced clearly for later retry.

#### Scenario: yt-dlp acquisition failure
- **WHEN** `yt-dlp` fails to retrieve/process audio for an article video
- **THEN** the pipeline SHALL record/log a clear failure reason for that article
- **AND** the article SHALL remain retryable in a subsequent run
