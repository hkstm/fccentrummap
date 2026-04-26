## 1. Video detection

- [x] 1.1 Extract embedded YouTube `video_id` from fetched article HTML
- [x] 1.2 Persist `video_id` with the raw article record (or equivalent acquisition queue metadata)

## 2. Audio acquisition

- [x] 2.1 Implement `yt-dlp`-based audio download for detected videos
- [x] 2.2 Prefer WAV output when available; allow fallback formats (`m4a`, `mp3`, `flac`, `ogg`, `webm`, `mp4`)
- [x] 2.3 Store downloaded audio temporarily on local disk before DB ingestion

## 3. SQLite blob storage

- [x] 3.1 Add schema for durable per-article audio blob storage
- [x] 3.2 Store `audio_blob` + metadata (`audio_format`, `mime_type`, `byte_size`, `youtube_url`, `video_id`)
- [x] 3.3 Ensure idempotent reruns skip articles that already have stored audio blobs

## 4. Verification

- [x] 4.1 Verify embedded-video articles produce stored audio blob rows
- [x] 4.2 Verify accepted output formats are recorded correctly in metadata
- [x] 4.3 Verify failed downloads are surfaced clearly and can be retried
