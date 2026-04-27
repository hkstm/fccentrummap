## Context

The scraper currently stores raw article HTML in `articles_raw`, but no normalized article text is persisted for later use. Downstream features that need article copy must repeatedly parse HTML, which is brittle and expensive to maintain. This change introduces durable extraction of cleaned main-content text via Trafilatura.

Constraints:
- Keep the current Go + SQLite stack; no new external services.
- Integrate with existing scraper/repository flow and idempotent schema initialization.
- Avoid introducing behavior that breaks existing scraping/export paths.

## Goals / Non-Goals

**Goals:**
- Persist queryable article text segments in a dedicated table linked to `article_raw_id`.
- Extract cleaned main-content text through a single Trafilatura-based path.
- Integrate extraction into scraper processing so text is stored automatically after successful fetch.
- Keep inserts idempotent and safe for reruns.

**Non-Goals:**
- General-purpose readability extraction across arbitrary HTML structures.
- NLP cleanup/summarization, language normalization, or semantic tagging.
- Frontend rendering or public API exposure of extracted text in this change.

## Implementation Update (2026-04-27)

This change was simplified to a single extraction path using **Trafilatura** for main-content extraction. The prior template-specific rule strategy was replaced to reduce maintenance burden and improve robustness for LLM-oriented downstream use.

## Decisions

1. **Split storage into extraction-run log + extracted content tables**
   - Decision: Create two related tables:
     - `article_text_extractions` (one row per extraction attempt per article) with at least `extraction_id`, `article_raw_id`, `extraction_mode`, `status`, `matched_count`, optional `error_message`, timestamps.
     - `article_text_contents` (0..N rows) with at least `text_content_id`, `extraction_id`, `article_raw_id`, `source_type`, `content`, timestamps.
   - Rationale: Separates operational observability (what happened) from content payload (what was extracted).
   - Alternatives considered:
     - Store extracted text JSON in `articles_raw` → rejected (harder to query, mixes raw and derived data).
     - Only store content rows and infer failures from emptiness → rejected (cannot distinguish no-match vs parser failure).

2. **Use Trafilatura as the single extraction mechanism**
   - Decision: Parse raw HTML with Trafilatura and use extracted main-content text as the canonical extraction output.
   - Rationale: More robust across page-template variations and better aligned with downstream LLM context preparation.
   - Alternatives considered:
     - Template-specific rule extraction only → rejected (high maintenance and brittle to template drift).
     - Full document text flattening → rejected (too much boilerplate noise).

3. **Treat low-content outputs as no-match**
   - Decision: If Trafilatura extraction returns empty/very short content (below threshold), classify as `no_match`.
   - Rationale: Prevents tiny boilerplate snippets from being treated as successful article extraction.
   - Alternatives considered:
     - Mark any non-empty output as matched → rejected (too noisy, poor signal quality).

4. **Record no-match and error outcomes explicitly**
   - Decision: Always insert an `article_text_extractions` row, even when no text is extracted or parsing fails.
   - Rationale: Provides an audit trail and supports drift monitoring when site formats change.
   - Alternatives considered:
     - Skip writes on no-match/error → rejected (loses critical diagnostics).

5. **Persist matched segments as individual rows with source metadata (no ordering metadata)**
   - Decision: For successful matches, store each extracted segment as one row and track source type; do not add sequence index in this phase.
   - Rationale: Preserves provenance while keeping schema minimal per current requirements.
   - Alternatives considered:
     - Single concatenated text blob per article → rejected (loses segment boundaries).
     - Add sequence metadata now → rejected as unnecessary for current use.

6. **Integrate extraction in scraper pipeline after raw HTML storage**
   - Decision: Run extraction during article processing once HTML is available, then write extraction-log row and content rows in one transaction.
   - Rationale: Ensures data is produced automatically and remains consistent across reruns.
   - Alternatives considered:
     - Separate backfill CLI only → rejected for initial rollout (adds operational overhead).

7. **Use replace-or-rebuild strategy per article for idempotency**
   - Decision: Before writing new results for an article, clear prior extraction-log/content rows for that article inside a transaction, then insert fresh rows.
   - Rationale: Prevents duplicates and keeps one authoritative latest extraction outcome per article.
   - Alternatives considered:
     - Upsert by content uniqueness rules → rejected (normalization collisions and noisy uniqueness semantics).

## Risks / Trade-offs

- **[Risk] Main-content extractor drift on website layout changes** → Mitigation: explicitly log `no_match` outcomes and monitor trend increases as drift signal.
- **[Risk] Short/boilerplate-only extraction treated as success** → Mitigation: enforce minimum extracted-character threshold before `matched`.
- **[Risk] Extra DB growth from extraction log + per-segment rows** → Mitigation: normalized text only and one latest-authoritative result per article.
- **[Risk] Rebuild strategy can temporarily remove rows if transaction handling is wrong** → Mitigation: perform delete+insert in one transaction and test rollback behavior.
- **[Trade-off] Main-content extraction may still include minor clutter** → Accepted for robust LLM-context capture; downstream summarization/cleanup can trim residual noise.

## Migration Plan

1. Extend repository schema init to create `article_text_extractions` and `article_text_contents` idempotently.
2. Add repository methods to replace/read latest extraction outcome and extracted segments by `article_raw_id`.
3. Implement extractor utility that runs Trafilatura and returns one outcome: `matched`, `no_match`, or `error`.
4. Wire extraction into scraper processing after raw HTML fetch/persist.
5. Persist extraction-log row and content rows transactionally; enforce one authoritative latest result per article on rerun.
6. Validate with representative FC Centrum pages, including matched/no-match/error scenarios.
7. Rollback: stop writing new rows (code rollback); existing tables can remain without affecting current consumers.

## Open Questions

- None currently.
