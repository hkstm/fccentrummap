# PR #1 review comment plan

PR: https://github.com/hkstm/fccentrummap/pull/1

## `gh` CLI usage I checked
- `gh pr view --help`
- `gh pr view 1 --repo hkstm/fccentrummap --json ...`
- `gh api repos/hkstm/fccentrummap/pulls/1/comments --paginate`

---

## Per-comment plan

### 2) Add language tag to fenced block (extract change design)
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021887
- **Reviewer comment:** markdown lint `MD040`.
- **Plan:** **Fix**.
- **How to fix:** change opening fence to ````text` in `openspec/changes/extract-spots-from-video-transcripts/design.md` for the transcript example block.

### 4) Missing `video_url` / `timestamp_seconds` in current schema
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021918
- **Reviewer comment:** repository schema lacks fields defined in change delta.
- **Plan:** **Reply only (already replied, acknowledged by reviewer bot)**.
- **Suggested reply (already used):** this is planned delta work under `openspec/changes/...`; canonical current requirements are in `openspec/specs/...`, so current schema is intentionally unchanged until `/opsx-apply extract-spots-from-video-transcripts`.

### 5) Missing fallback behavior when transcript unavailable
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021924
- **Reviewer comment:** no mandatory behavior for no-transcript case.
- **Plan:** **Fix**.
- **How to fix:** add scenario to `video-spot-extraction/spec.md`:
  - mark article extraction state (`no_transcript`),
  - skip spot extraction for that article,
  - continue batch,
  - emit retryable metric/event.

### 6) Video ID handoff contract underspecified
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021928
- **Reviewer comment:** "make identifier available" is vague.
- **Plan:** **Fix**.
- **How to fix:** in `extract-spots-from-video-transcripts/specs/web-scraper/spec.md`, explicitly require normalized `video_id` storage (nullable string, canonical 11-char YouTube ID) on scraped article metadata for transcript stage.

### 7) Geocoding orchestration described but not implemented (old scrape change)
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021935
- **Reviewer comment:** design overstates implementation.
- **Plan:** **Reply (resolved by archive narrowing)**.
- **Suggested reply:** this was addressed by narrowing + archiving `scrape-spots-to-sqlite` to raw-ingestion foundation scope; geocoding orchestration is intentionally deferred to successor transcript-first work.

### 8) Archived spec says “skip fetching” duplicates but code fetches then dedupes on insert
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021940
- **Reviewer comment:** requirement mismatch.
- **Plan:** **Fix doc**.
- **How to fix:** in archived `.../specs/web-scraper/spec.md`, change scenario wording from "skip fetching" to "skip duplicate insert/storage" to reflect implemented behavior.

### 9) Verification checklist incomplete (old scrape tasks)
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021944
- **Reviewer comment:** pending verification tasks.
- **Plan:** **Reply**.
- **Suggested reply:** this comment targeted pre-archive `scrape-spots-to-sqlite`; tasks were rewritten when narrowing/archive. Remaining end-to-end checks belong to successor changes and will be completed when those changes are applied.

### 10) Handle `file.Close()` error in export command
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021950
- **Reviewer comment:** close errors ignored.
- **Plan:** **Fix**.
- **How to fix:** replace `defer file.Close()` with deferred closure checking close error and logging/failing.

### 11) Upgrade vulnerable `antchfx/xpath` to v1.3.6
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021958
- **Reviewer comment:** `v1.3.5` vulnerable.
- **Plan:** **Fix**.
- **How to fix:** run in `scraper/`:
  - `go get github.com/antchfx/xpath@v1.3.6`
  - `go mod tidy`
  - commit updated `go.mod`/`go.sum`.

### 12) Add timeout to geocoder API call
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021959
- **Reviewer comment:** `context.Background()` can hang.
- **Plan:** **Fix**.
- **How to fix:** use `context.WithTimeout(..., 10*time.Second)` in `Geocode`, `defer cancel()`, and pass `ctx` to `g.client.Geocode`.

### 14) Use transaction in `LinkArticleSpots`
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021985
- **Reviewer comment:** partial commits possible.
- **Plan:** **Fix**.
- **How to fix:** wrap loop in transaction (`Begin`/`tx.Exec`/`Commit`, rollback on error).

### 15) Fail fast if `data-max-page` missing/invalid
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074021992
- **Reviewer comment:** currently can silently skip paginated pages.
- **Plan:** **Fix**.
- **How to fix:** validate parsed max page after first visit; if invalid/missing, return error instead of silently continuing.

### 16) Add Colly `OnError` callback for async fetch failures
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074022004
- **Reviewer comment:** network failures can be dropped.
- **Plan:** **Fix**.
- **How to fix:** add `c.OnError(...)` in `FetchAndStoreArticles` to set `fetchErr` (first error wins) and keep existing `OnResponse` path.

### 17) Archived spec still includes unimplemented parse requirements
- **Comment link:** https://github.com/hkstm/fccentrummap/pull/1#discussion_r3074080424
- **Reviewer comment:** archived spec overstates scope.
- **Plan:** **Fix doc**.
- **How to fix:** in archived `.../specs/web-scraper/spec.md`, remove/relax mandatory parse requirements (author/spot parsing), and explicitly defer to successor change `extract-spots-from-video-transcripts`.

---

## Suggested execution order
1. **Spec/doc fixes first:** #2, #3, #5, #6, #8, #17
2. **Small reliability/security code fixes:** #10, #11, #12, #13, #14, #15, #16
3. **Post replies/closures:** #1, #4, #7, #9

This keeps narrative coherent: docs/spec truthfulness first, then implementation hardening, then thread replies.
