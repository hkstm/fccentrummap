## Context

The desired extraction direction for fccentrummap is now transcript-first: detect the embedded YouTube video for each article, generate a transcript, and run an LLM extractor on that transcript to recover spots, addresses, and timestamps.

Any earlier article-text extraction implementation should be considered deprecated and not a required precursor for this change. This design should be read as a replacement path, not as an incremental extension of code that must still exist.

**Target flow:**
1. Fetch article HTML → store in `articles_raw`
2. Detect embedded YouTube video ID from article HTML
3. Download and parse transcript for that video
4. Send transcript segments to the LLM extractor
5. LLM returns `{author_name, spots: [{name, address, timestamp_seconds}]}`
6. Geocode addresses → store in `spots` table

**Constraint:** We need timestamps to support future map features (e.g., clicking a marker opens the YouTube video at that spot's timestamp).

## Goals / Non-Goals

**Goals:**
- Extract spots from embedded videos using a transcript-first pipeline
- Obtain timestamps for each spot to enable YouTube deep-linking (`?t=123s`)
- Minimize API costs (prefer free YouTube auto-captions over paid Whisper API)
- Produce a structured LLM extraction result directly from transcript content

**Non-Goals:**
- Real-time transcription (batch processing is fine, videos are already published)
- Multi-language support beyond Dutch (fccentrum.nl is Dutch-only)
- Video analysis beyond audio/captions (no computer vision for reading street signs)
- Handling private/unlisted videos (all fccentrum videos are public)

## Decisions

### Decision 1: Use yt-dlp for transcript extraction

**Choice:** Shell out to `yt-dlp --write-auto-sub --sub-lang nl --skip-download` to download YouTube's auto-generated Dutch subtitles in SRT format.

**Rationale:**
- **Free:** YouTube auto-captions are free, Whisper API costs ~$0.006/minute ($0.18 for 30min video)
- **Fast:** yt-dlp downloads pre-generated captions in <1 second (vs. Whisper which transcribes audio in ~15% realtime)
- **Timestamps included:** SRT format has built-in timestamps for every subtitle chunk
- **Mature tooling:** yt-dlp is actively maintained, handles rate limits, supports all YouTube embed formats

**Alternatives considered:**
- **Whisper API (OpenAI):** Better accuracy but costs money and requires audio extraction step. Reserve as fallback for videos without auto-captions.
- **Gemini multimodal video API:** Can process video files directly but has upload size limits (unclear if 30min videos work) and potentially higher costs. Overkill for our use case.
- **YouTube Data API transcript endpoint:** Requires OAuth, fragile (caption availability not guaranteed), less reliable than yt-dlp.

**Trade-off:** Auto-captions may mangle Dutch street names/addresses (e.g., "Elandsgracht" → "eh lands grat"). Mitigation: the LLM extractor should normalize likely Dutch place names and addresses from timestamped transcript context.

### Decision 2: Build or regenerate a transcript-based extractor

**Choice:** Implement an extractor that accepts parsed transcript segments with timestamps and returns structured spots.

Example shape:
`ExtractFromTranscript(url string, transcript *Transcript)`

**Rationale:**
- **Matches current direction:** transcript is the intended extraction source
- **Clean replacement path:** does not assume an earlier article-text extractor still exists in the repo
- **Timestamp extraction:** the extractor can map each spot directly to a transcript time window

**Prompt structure:**
```
Extract author and recommended spots from this timestamped Dutch video transcript.

Video transcript:
[00:00:15] In De Spots van neemt Dai Carter...
[00:01:23] Het eerste plekje is A'DAM Music School...
[00:02:45] Dan gaan we naar café 't Molentje op de Prinsengracht...

Return JSON: {author_name, spots: [{name, address, timestamp_seconds}]}
```

**Alternatives considered:**
- **No LLM extraction, only regex/heuristics:** likely too brittle for noisy auto-captions
- **Use article text as co-input:** possible later, but not required for the initial transcript-first implementation

### Decision 3: Detect YouTube embeds during scraper phase

**Choice:** Add YouTube video ID extraction to the scraper's `FetchAndStoreArticles` function. Parse HTML for:
- `<iframe src="https://www.youtube.com/embed/{videoId}"`
- `<div class='flying-press-youtube' data-src='...embed/{videoId}'>`

Store video ID in a new `articles_raw.video_id TEXT` column.

**Rationale:**
- **Single-pass HTML parsing:** Already fetching and storing HTML, minimal overhead to extract video ID
- **Enables conditional transcript download:** Only call yt-dlp if `video_id IS NOT NULL`
- **Separates concerns:** Scraper handles HTML → video ID, separate transcript module handles video ID → SRT

**Alternatives considered:**
- **Extract video ID in extractor phase:** Requires re-parsing HTML, duplicates scraping logic
- **Always attempt yt-dlp (let it fail):** Wasteful for articles without embedded videos, adds 1-2s latency per article

### Decision 4: Store timestamps in spots table

**Choice:** Add `video_url TEXT NULL` and `timestamp_seconds INTEGER NULL` columns to `spots` table.

**Rationale:**
- **Future-proof for map markers:** Clicking a spot can open `https://youtube.com/watch?v={videoId}&t={timestamp}s`
- **Nullable for text-based spots:** Text-only articles won't have these fields, keeps schema backward-compatible
- **Single source of truth:** Spot ↔ timestamp lives in the same row

**Schema change:**
```sql
ALTER TABLE spots ADD COLUMN video_url TEXT;
ALTER TABLE spots ADD COLUMN timestamp_seconds INTEGER;
```

**Alternatives considered:**
- **Separate `spot_timestamps` table:** Over-normalized for a simple 1:1 relationship, adds join complexity
- **Store timestamp as TEXT (e.g., "1:23"):** Harder to query/sort, loses precision (SRT has centisecond granularity)

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| **Auto-caption quality for Dutch addresses** | LLM normalizes noisy caption text into likely place names and addresses. If quality is unacceptable, fall back to Whisper API for that video (manual flag or automatic retry). |
| **yt-dlp binary dependency** | Check for yt-dlp at startup, fail fast with clear error message. Document installation (`brew install yt-dlp` / `apt install yt-dlp`). Could bundle yt-dlp binary in future. |
| **YouTube changes embed format** | Regex for video ID extraction may break. Add test coverage for known embed formats. yt-dlp handles YouTube API changes upstream. |
| **Videos without auto-captions** | yt-dlp returns error → log warning, mark article as FAILED in `scrape_log`, and skip extraction for that article. Future: retry with Whisper. |
| **Timestamp alignment issues (LLM hallucinates timestamps)** | Validate timestamps against transcript duration (reject if `timestamp > video_length`). Include clear examples in LLM prompt to guide timestamp extraction. |
| **Performance regression (2-5s per video)** | Acceptable for batch scraping (43 articles × 3s = ~2 minutes total). If unacceptable, parallelize yt-dlp calls or pre-download all transcripts in bulk. |

## Migration Plan

**Schema migration:**
```sql
-- Add new columns (non-breaking, nullable)
ALTER TABLE spots ADD COLUMN video_url TEXT;
ALTER TABLE spots ADD COLUMN timestamp_seconds INTEGER;
ALTER TABLE articles_raw ADD COLUMN video_id TEXT;
```

**Deployment:**
1. Deploy schema changes (idempotent `ALTER TABLE IF NOT EXISTS`)
2. Deploy code with transcript download + transcript-based LLM extraction
3. Backfill existing articles: Re-scrape video-only articles (status=PENDING or manually trigger re-extraction)

**Rollback:**
- Code rollback: Safe (new columns ignored by old code)
- Schema rollback: `ALTER TABLE spots DROP COLUMN video_url, timestamp_seconds;` (loses timestamp data but spots remain)

## Open Questions

1. **Whisper API fallback:** Should we auto-retry with Whisper if yt-dlp fails, or require manual intervention? Leaning toward manual flag to avoid surprise API costs.
2. **Timestamp precision:** SRT has centisecond granularity (`00:01:23,450`). Do we store as `INTEGER` seconds (loses precision but simpler) or `REAL` for fractional seconds? Leaning toward seconds as YouTube `?t=` parameter only accepts whole seconds.
3. **Multi-video articles:** If an article embeds multiple videos, which one do we use? Current assumption: first video embed. Could revisit if we encounter articles with multiple spot videos.
