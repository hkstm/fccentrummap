## Context

The static export currently emits `spots` and `presenters` for the frontend map UI. `presenters` are sorted alphabetically, which is deterministic but does not reflect the editorial recency of the FC Centrum "Spots van" articles. The fetched article HTML already contains publish metadata (`article:published_time` and JSON-LD `datePublished`), and validation against the current SQLite dataset showed all fetched article HTML rows include matching publish timestamps.

The frontend should not need extra fields to order the filter. Article publication time should be persisted in SQLite as source metadata, and export should use that stored value for ordering while keeping the JSON contract compact and compatible.

## Goals / Non-Goals

**Goals:**
- Add dedicated SQLite storage for article publication time.
- Populate and backfill article publication time from metadata already present in fetched HTML.
- Sort exported presenters by the newest associated article publish timestamp, descending.
- Preserve the existing presenter JSON shape (`{ presenterName }`) without exposing publish timestamps.
- Keep export output deterministic with stable tie-breakers.
- Cover publication-time storage and ordering behavior with repository/export tests.

**Non-Goals:**
- Changing frontend filter sorting logic or adding client-side date parsing.
- Intentionally changing exported `spots` array ordering; the existing deterministic spot sort should remain unchanged.
- Re-scraping articles solely to populate publish metadata.

## Decisions

1. **Persist publication time on `article_sources`**
   - **Decision:** Add a nullable `published_at` timestamp column to `article_sources` and treat it as article-source metadata.
   - **Rationale:** Publication date belongs to the article source rather than an individual fetch, transcription, spot, or presenter. Persisting it avoids reparsing HTML during every export and makes ordering explicit in storage without changing the frontend JSON payload.
   - **Alternatives considered:**
     - Store on `article_fetches`: simple to populate when fetching, but publication time is not fetch-specific and would be tied to latest-fetch state.
     - Parse during export only: avoids schema work, but hides a durable data attribute in export logic and repeatedly parses HTML.
     - Add dates to exported presenters: makes frontend sorting possible but violates the desired output contract.

2. **Populate and backfill from fetched article HTML**
   - **Decision:** Parse `article:published_time` from fetched HTML when article fetches are stored, and backfill `article_sources.published_at` for existing compatible databases from current `article_fetches.html`.
   - **Rationale:** All current fetched article HTML has publish metadata, and existing databases need the column populated before export ordering can be meaningful.
   - **Alternatives considered:**
     - Require a full re-scrape: unnecessary because stored HTML already contains the metadata.
     - Manual SQL updates: brittle and not repeatable.

3. **Require publish metadata for exportable article rows**
   - **Decision:** Treat missing or unparseable article publication time for exportable rows as an export/backfill error rather than silently falling back to alphabetical placement.
   - **Rationale:** Recency ordering is only trustworthy if every exported presenter has reliable article publication metadata; failing loudly prevents subtle stale or misleading ordering.
   - **Alternatives considered:**
     - Sort undated presenters after dated presenters: deterministic, but hides data quality issues.
     - Preserve alphabetical order entirely when any date is missing: avoids errors, but defeats the purpose of recency ordering.

4. **Track latest publish time per presenter**
   - **Decision:** While scanning export rows, update an internal map from presenter name to max publish timestamp.
   - **Rationale:** A presenter can be associated with multiple articles; the requirement is based on their latest article.
   - **Alternatives considered:**
     - Sort by first encountered article: dependent on query order and not semantically tied to recency.
     - Sort by latest exported spot timestamp: video timestamps are not publication recency.

5. **Keep deterministic tie-breaker ordering**
   - **Decision:** Sort presenters by latest publish time descending, then presenter name ascending when timestamps are equal.
   - **Rationale:** Ensures stable output for equal timestamps.
   - **Alternatives considered:**
     - Preserve query order for ties: not deterministic enough.

6. **Prefer `article:published_time` with optional JSON-LD fallback**
   - **Decision:** Use `article:published_time` as the primary source and `datePublished` as a fallback if useful.
   - **Rationale:** Both are present and match today; `article:published_time` is simple to parse from meta tags.
   - **Alternatives considered:**
     - Full HTML parser dependency: unnecessary for known metadata shape.
     - Regex-only JSON-LD parsing as primary: less direct than the explicit OpenGraph article metadata.

## Risks / Trade-offs

- **[Risk] Publish metadata format changes in future HTML** → **Mitigation:** Fail backfill/export with actionable diagnostics and add tests for the supported metadata forms.
- **[Risk] Existing databases lack the new column/value** → **Mitigation:** Add idempotent schema evolution for compatible v2-era databases and backfill from stored `article_fetches.html`; fail if required metadata cannot be parsed.
- **[Risk] Parsing dates while storing fetches adds coupling to HTML shape** → **Mitigation:** Keep parsing localized in repository/fetch persistence helpers and do not expose dates in exported models.
- **[Risk] Multiple articles for one presenter include duplicate/alias URLs** → **Mitigation:** Sorting uses max publish time per presenter and remains independent of duplicate spot cleanup.
- **[Trade-off] Schema evolution adds implementation complexity** → **Mitigation:** Use a nullable column and idempotent `ALTER TABLE`/backfill path rather than a destructive migration.
