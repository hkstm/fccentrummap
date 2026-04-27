## Context

The current workflow to test transcription requires manual one-off steps: query SQLite for an audio blob, export to a local file, call Murmel with `curl`, and save the JSON response manually. This is repetitive and hard to reproduce.

Existing data already includes `article_audio_sources` rows with the source audio blob and metadata. The codebase already manages SQLite schema initialization through the repository layer (`modernc.org/sqlite`). We need an automated CLI path that:
1) picks an audio source row (explicit ID or latest),
2) sends it to Murmel using configured API auth,
3) stores the transcription response linked to the originating audio row.

Constraints:
- Keep existing scraper/audio ingestion behavior unchanged.
- Reuse existing SQLite initialization patterns instead of introducing ad-hoc migration tooling.
- Handle Murmel auth via environment configuration and support operational debugging (status, error payloads).

## Goals / Non-Goals

**Goals:**
- Add a Go CLI command to transcribe one `article_audio_sources` row using Murmel.
- Support selection by `--audio-source-id` and fallback to latest row when omitted.
- Persist Murmel response JSON in a new DB table with FK reference to `article_audio_sources.audio_source_id`.
- Persist minimal request/response metadata (language, HTTP status, payload size, created timestamp, optional error message).
- Keep command output clear for local testing (selected ID, Murmel status, saved record ID).

**Non-Goals:**
- Batch transcription of all rows.
- Replacing existing audio-acquisition pipeline.
- Introducing async job orchestration/queues.
- Normalizing Murmel response schema into many relational fields; raw JSON remains canonical stored payload.

## Decisions

1. **Introduce a dedicated transcription results table with uniqueness**
   - Decision: Add table `article_audio_transcriptions` with FK to `article_audio_sources` and a UNIQUE constraint on `(audio_source_id, provider, language)`.
   - Rationale: Prevents accidental duplicate rows for the same source/provider/language while still allowing intentional re-runs via upsert/replace behavior.
   - Proposed columns:
     - `transcription_id` INTEGER PK
     - `audio_source_id` INTEGER NOT NULL FK (`article_audio_sources.audio_source_id`)
     - `provider` TEXT NOT NULL (default `murmel`)
     - `language` TEXT NOT NULL
     - `http_status` INTEGER NOT NULL
     - `response_json` TEXT NOT NULL
     - `response_byte_size` INTEGER NOT NULL
     - `error_message` TEXT NULL
     - `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   - Constraint/indexes:
     - `UNIQUE(audio_source_id, provider, language)`
     - Optional expression indexes on frequently queried JSON paths (e.g. `json_extract(response_json, '$.status')`) when query patterns are known.
   - Alternatives considered:
     - Store transcription on `article_audio_sources` directly → rejected because it couples mutable transcription state to immutable source audio.
     - Allow unlimited duplicate history → rejected based on user requirement to enforce uniqueness.

2. **Store JSON as canonical TEXT for interoperability and flexible JSON1 querying**
   - Decision: Store Murmel payload in `response_json` as canonical JSON TEXT, normalized at write time (`json(?)`) and validated (`CHECK(json_valid(response_json))`).
   - Rationale: Per SQLite JSON1 docs, SQLite stores JSON as ordinary text and JSON functions/operators work directly on JSON text; this maximizes compatibility with tooling, easy export, and path-based querying.
   - Performance approach:
     - Accept slightly higher write-time transform cost (canonicalization/validation).
     - Optimize reads using JSON path expression indexes and/or generated columns for hot fields.
   - Alternatives considered:
     - Store as raw BLOB/JSONB only → rejected for now due to reduced portability/inspectability and stronger coupling to SQLite-internal binary representation.

3. **CLI selection strategy: explicit ID first, otherwise latest**
   - Decision: `transcribe-audio` accepts optional `--audio-source-id`; if missing, query latest by highest `audio_source_id` with non-empty blob.
   - Rationale: Matches current testing workflow and keeps command ergonomic.
   - Alternatives considered:
     - Require ID always → rejected as too cumbersome for quick testing.
     - Choose by latest `created_at` only → rejected due to potential timestamp inconsistencies; `audio_source_id` is deterministic.

4. **Murmel API integration uses `X-API-Key` header and multipart form**
   - Decision: Send request as multipart form with `audio` file part and `language` field; authenticate via `X-API-Key` from env.
   - Rationale: Matches observed working API contract and avoids auth mismatch.
   - Alternatives considered:
     - Bearer auth header → rejected (known 401 behavior in current environment).

5. **Persist response for both success and failure**
   - Decision: Always persist response body and metadata; record `error_message` for transport/HTTP failures where available.
   - Rationale: Enables auditability and easier retry/debug loops.
   - Alternatives considered:
     - Save only 2xx responses → rejected because failed responses are often most valuable diagnostically.

6. **Add export CLI support for both audio and transcription JSON artifacts**
   - Decision: Add export commands that can write either audio payload or transcription JSON payload to the `data/` directory by explicit ID.
   - Rationale: Supports fast local iteration and manual inspection without ad-hoc SQL/curl steps.
   - Scope:
     - Export audio by `audio_source_id`.
     - Export transcription JSON by `transcription_id`.
     - Deterministic output filenames including IDs and detected extensions.

7. **Schema migration approach: idempotent create-on-startup**
   - Decision: Extend repository schema initialization with `CREATE TABLE IF NOT EXISTS article_audio_transcriptions ...` and unique index/constraint creation.
   - Rationale: Aligns with existing project schema strategy and avoids adding a separate migration system.
   - Alternatives considered:
     - Versioned migration framework → rejected as unnecessary scope for this incremental change.

## Risks / Trade-offs

- **[Large response payload growth]** Transcription JSON blobs can increase DB size quickly → **Mitigation:** store byte size metadata and keep one-row-per-run model for optional cleanup policies later.
- **[Network/API flakiness]** Murmel downtime or transient failures can produce incomplete workflows → **Mitigation:** explicit non-zero CLI exit codes plus persisted failure responses for retries.
- **[Credential misconfiguration]** Missing/invalid API key leads to frequent failures → **Mitigation:** validate env key at command start and return actionable error text.
- **[Schema drift across local DBs]** Older DB files may lack new table → **Mitigation:** idempotent schema initialization at startup before command execution.
- **[Memory pressure]** Building multipart payload from large audio blobs may be expensive → **Mitigation:** stream blob to temp file/request writer where possible instead of duplicating large byte slices.

## Migration Plan

1. Add new table creation SQL to repository initialization.
2. Add repository methods:
   - get audio source by ID,
   - get latest audio source,
   - insert transcription result row.
3. Add CLI command wiring and flags (`--audio-source-id`, `--language` default `nl`, optional `--db-path`).
4. Implement Murmel client helper using `X-API-Key` and multipart upload.
5. Validate command against existing `data/spots.db`.
6. Rollback strategy: if needed, disable command usage and ignore new table; existing pipeline remains unaffected.

## Open Questions

- None for this iteration. Follow-up refinements (batch mode, retention policies, and additional JSON indexes) can be evaluated after first implementation.