## ADDED Requirements

### Requirement: Transcript extraction uses YouTube auto-captions first
The extraction pipeline SHALL obtain transcript data for embedded YouTube videos using YouTube auto-generated subtitles before considering paid transcription fallbacks.

#### Scenario: Auto-captions available
- **WHEN** an article has an embedded YouTube video with Dutch auto-captions available
- **THEN** the pipeline SHALL download and parse those subtitles as the transcript source

#### Scenario: Auto-captions unavailable
- **WHEN** an embedded YouTube video does not provide usable auto-captions
- **THEN** the pipeline SHALL report transcript acquisition failure clearly so the article can be retried or handled by a fallback path later

### Requirement: Transcript segments preserve timestamps
Transcript extraction SHALL preserve subtitle timing information in a structured form suitable for downstream spot extraction.

#### Scenario: Parsed subtitle segment
- **WHEN** a subtitle file is parsed successfully
- **THEN** each transcript segment SHALL retain timestamp information that can be mapped to a later `timestamp_seconds` output value
