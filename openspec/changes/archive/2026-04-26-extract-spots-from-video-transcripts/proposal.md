## Why

Before we build transcript extraction and spot parsing, we need a reliable ingestion stage that turns each article's embedded YouTube video into a durable local audio artifact. Storing audio directly in SQLite as a blob gives us a reproducible, offline-friendly input dataset for later processing.

This change intentionally narrows scope to **article -> video detection -> audio download -> SQLite blob storage**.

## What Changes

- Detect embedded YouTube video IDs from fetched article HTML
- Download audio for detected videos using `yt-dlp` (via local temporary file)
- Store downloaded audio payloads as BLOBs in SQLite, with format and MIME metadata
- Prefer WAV output when available, while accepting fallback formats (`m4a`, `mp3`, `flac`, `ogg`, `webm`, `mp4`)

## Capabilities

### New Capabilities

- `youtube-audio-acquisition`: Acquire article-linked YouTube audio with `yt-dlp` and persist binary audio payloads for downstream processing

### Modified Capabilities

- `web-scraper`: Detect embedded YouTube IDs during article fetch/storage
- `sqlite-storage`: Add durable storage for per-article audio blobs + audio metadata

## Impact

- **Database schema:** new audio-source storage table (or equivalent) with BLOB payloads and metadata
- **New dependency:** `yt-dlp` (and `ffmpeg` when format conversion is required, e.g. WAV extraction)
- **Pipeline shape:** adds an explicit audio-ingestion stage after raw HTML ingestion
- **Storage footprint:** database size increases due to binary audio blobs
