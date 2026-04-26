## Context

We are narrowing this change to a single ingestion concern: convert embedded YouTube videos referenced by scraped articles into locally persisted audio blobs in SQLite.

This is a foundation step for future transcript/extraction work, but this change itself does not implement transcript parsing, LLM extraction, geocoding, or spot writes.

**Target flow:**
1. Fetch/store article HTML in `articles_raw`
2. Detect embedded YouTube `video_id`
3. Download audio for that video using `yt-dlp` into a local temp file
4. Read bytes from temp file and store as SQLite BLOB with metadata
5. Mark acquisition success/failure for retryability

## Goals / Non-Goals

**Goals:**
- Reliably detect YouTube embeds from article HTML
- Acquire audio with `yt-dlp` for detected videos
- Persist audio payloads as SQLite BLOBs with enough metadata for downstream consumers
- Keep the ingestion rerunnable and idempotent

**Non-Goals:**
- Transcript generation/parsing
- LLM extraction of spots/authors/addresses
- Timestamp extraction and deep-link metadata
- Frontend/export contract changes tied to transcript-derived data

## Decisions

### 1) Store audio in a dedicated SQLite table
**Choice:** Use a dedicated table keyed by `article_raw_id`.

```sql
CREATE TABLE IF NOT EXISTS article_audio_sources (
    audio_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
    article_raw_id INTEGER NOT NULL UNIQUE REFERENCES articles_raw(article_raw_id),
    video_id TEXT NOT NULL,
    youtube_url TEXT NOT NULL,
    audio_format TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    audio_blob BLOB NOT NULL,
    byte_size INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Rationale:**
- Avoids bloating `articles_raw` with binary payload columns
- Gives a clear one-row-per-article audio artifact
- Keeps downstream reads simple (`JOIN` by `article_raw_id`)

### 2) Acquisition strategy with yt-dlp
**Choice:** Use `yt-dlp` to download audio to a local temp file, then ingest file bytes into SQLite.

**Rationale:**
- `yt-dlp` is robust for YouTube handling
- Temp-file flow is straightforward and observable
- SQLite stores final canonical artifact after ingestion

### 3) Audio format policy
**Choice:** Prefer WAV output when available; otherwise accept source/container outputs among:
`m4a`, `mp3`, `flac`, `ogg`, `webm`, `mp4`.

**Rationale:**
- WAV is convenient for later DSP/transcription tooling
- Fallback formats are constrained to the practical ffmpeg-supported subset for this project setup

### 4) Idempotency and retries
**Choice:** If an article already has an audio row in `article_audio_sources`, skip re-acquisition unless explicitly forced.

**Rationale:**
- Avoid repeated network/download work
- Keep reruns safe and predictable

## Risks / Trade-offs

- **Database growth from BLOB storage:** mitigated by one-row-per-article policy and optional future compression/archive strategy.
- **Dependency availability (`yt-dlp`, `ffmpeg`):** mitigated by startup checks and clear error messages.
- **Format variability:** mitigated by storing explicit `audio_format` + `mime_type` metadata.
- **Download failures / unavailable videos:** mitigated by clear failure status and retryable workflow.

## Migration Plan

1. Add schema for `article_audio_sources`
2. Add repository methods for insert/get/exists for audio sources
3. Implement acquisition stage (`video_id` -> temp audio file -> BLOB insert)
4. Validate rerun behavior (skip already-stored audio rows)

**Rollback:**
- Code rollback is safe (table remains unused by older code)
- Schema rollback can drop `article_audio_sources` if data retention is not required

## Open Questions

1. Should failed acquisitions be tracked in `articles_raw.status`, a new status column on `article_audio_sources`, or both?
2. Do we want to retain local temp/source audio files after BLOB insertion for debugging?
3. Should we set a maximum blob size guardrail per article to avoid pathological downloads?
