## ADDED Requirements

### Requirement: Transcription stage supports unified scrape entrypoint
Transcription behavior SHALL be invocable via unified scrape stage command semantics while preserving Murmel contract and persistence behavior.

#### Scenario: Unified stage invocation
- **WHEN** a user runs `transcribe-audio` through the unified scrape entrypoint
- **THEN** the stage SHALL transcribe selected stored audio through Murmel
- **AND** it SHALL persist canonical transcription result rows in SQLite

### Requirement: Transcription stage supports deterministic file-mode handoff
The transcription stage SHALL support explicit file-mode input and deterministic output artifact naming when run outside SQLite mode.

#### Scenario: File-mode transcription invocation
- **WHEN** a user runs transcription with `--io file` and explicit `--in`
- **THEN** the stage SHALL process that explicit input artifact
- **AND** it SHALL emit deterministically named output artifact(s)
