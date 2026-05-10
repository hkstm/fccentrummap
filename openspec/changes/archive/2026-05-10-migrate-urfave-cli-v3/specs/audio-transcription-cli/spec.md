## MODIFIED Requirements

### Requirement: Transcription stage supports unified scrape entrypoint
Transcription behavior SHALL be invocable via urfave/cli v3 unified scrape stage command semantics while preserving Murmel contract and persistence behavior.

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

## ADDED Requirements

### Requirement: Transcription CLI MAY prefer idiomatic urfave/cli v3 behavior
The transcription CLI SHALL permit idiomatic urfave/cli v3 command/flag behavior and SHALL NOT be required to retain legacy stdlib-flag invocation quirks.

#### Scenario: Legacy invocation quirk removed
- **WHEN** a legacy invocation form exists only due to prior stdlib-flag parsing behavior
- **THEN** the urfave/cli v3 command MAY reject that form
- **AND** the in-repo CLI documentation SHALL define the supported invocation pattern
