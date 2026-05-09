# youtube-audio-acquisition Specification

## Purpose
Define the canonical requirements for acquiring audio from embedded YouTube videos so the FC Centrum pipeline can produce durable, replayable inputs for transcript generation and downstream spot extraction. This specification is for scraper/pipeline contributors and downstream processing stages, and it standardizes how detected `video_id` inputs are converted into stored audio artifacts plus metadata while handling runtime constraints such as `yt-dlp`/`ffmpeg` availability, format fallback behavior, and clear retryable failure reporting.
## Requirements
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

### Requirement: Audio acquisition stage supports unified I/O contract
The audio acquisition stage SHALL support unified CLI I/O mode behavior with SQLite as default and explicit failure for unsupported combinations.

#### Scenario: Acquire audio in sqlite mode
- **WHEN** the stage runs in default sqlite mode
- **THEN** it SHALL read candidate records from SQLite and persist acquired audio metadata/blob to SQLite

#### Scenario: Unsupported invocation mode
- **WHEN** the stage is invoked in a mode not implemented for that command
- **THEN** it SHALL fail with a non-zero status and actionable guidance

