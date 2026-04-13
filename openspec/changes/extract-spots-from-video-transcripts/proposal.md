## Why

Some fccentrum.nl "De Spots van" articles contain the useful spot information in embedded YouTube videos rather than in the article body. We want the extraction pipeline to use a transcript generated from the embedded video as the primary extraction input.

Note: this change should be treated as the source of truth for extraction direction. The repository no longer needs to preserve or extend the earlier article-text extraction implementation.

## What Changes

- Add YouTube transcript extraction using yt-dlp to download auto-generated Dutch subtitles
- Extend the scraping pipeline to detect embedded YouTube videos in articles
- Add an LLM extraction step that reads the generated video transcript and returns structured spots
- Extract timestamps from subtitles to enable deep-linking to specific spots in the video
- Store video URLs and spot timestamps in the database for future map marker links

## Capabilities

### New Capabilities

- `youtube-transcript-extraction`: Download and parse YouTube auto-generated subtitles (SRT format) for fccentrum.nl embedded videos, with fallback to Whisper API if auto-captions are unavailable or poor quality
- `video-spot-extraction`: Extract spot names, addresses, and timestamps from the generated video transcript using an LLM
- `video-metadata-storage`: Store video URLs, timestamps, and transcript references in the database to support future features like clickable map markers that deep-link to the relevant video moment

### Modified Capabilities

- `web-scraper`: Detect and extract YouTube video IDs from article HTML (embedded iframes/shortcodes) during the fetch phase
- `sqlite-storage`: Add `video_url` and `timestamp` columns to the `spots` table, update schema to link spots to specific moments in source videos

## Impact

- **Database schema**: New columns in `spots` table (`video_url TEXT`, `timestamp INTEGER` for seconds offset)
- **New dependency**: `yt-dlp` binary (installed via package manager or bundled)
- **Extractor changes**: Add or regenerate an extractor that operates on transcript content produced from the embedded video
- **Scraper changes**: Parse YouTube embeds during article fetch phase
- **API costs**: Potential Whisper API usage if auto-captions fail (fallback only, ~$0.006/minute)
- **Performance**: Additional ~2-5 seconds per video-only article for yt-dlp download + parsing
