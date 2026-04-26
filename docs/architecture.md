# Architecture

This document summarizes the implemented architecture.

## Canonical source

For normative behavior and architectural rules, use:

- `openspec/specs/project-layout/spec.md`
- `openspec/specs/web-scraper/spec.md`
- `openspec/specs/sqlite-storage/spec.md`
- `openspec/specs/static-data/spec.md`
- `openspec/specs/frontend-portability/spec.md`

## High-level pipeline

```text
scraper -> data/spots.db -> viz/public/data/spots.json -> viz frontend
```

## Current implementation shape

- `scraper/cmd/scraper` crawls category pages, fetches article HTML, stores raw inputs, and reports pending work
- `scraper/cmd/export` reads SQLite and writes frontend JSON
- `viz/` is reserved for frontend work that consumes `/data/spots.json`

## Current scraper processing flow (implemented)

```mermaid
flowchart TD
    A[Start scraper/cmd/scraper main] --> B[Open SQLite repo]
    B --> C[InitSchema]
    C --> D{article-url flag provided?}

    D -- Yes --> D1[Use single URL from flag]
    D -- No --> E[CrawlArticleURLs]

    subgraph Crawl phase
      E --> E1[Visit base category page]
      E1 --> E2[Read data-max-page]
      E2 --> E3[Visit page 2..N]
      E3 --> E4[Extract article links]
      E4 --> E5[Deduplicate URLs in-memory via seen map]
    end

    D1 --> F[FetchAndStoreArticles urls]
    E5 --> F

    subgraph Fetch + store phase
      F --> G{For each URL}
      G --> H[HTTP fetch via Colly Visit]
      H --> I[OnResponse: extract optional YouTube video_id]
      I --> J[InsertArticleRaw url, html, video_id]
      J --> K[SQL: INSERT OR IGNORE into articles_raw]
      K --> L{URL already exists?}
      L -- No --> M[Row inserted status=PENDING]
      L -- Yes --> N[Insert ignored]
      M --> G
      N --> G
    end

    G --> O[AcquireAndStoreAudio context, repo, nil]
    O --> P[Store article_audio_sources via INSERT OR IGNORE]
    P --> Q[Exit]

    X[[Important current behavior]] -.-> H
    X -.-> N
    X["No pre-fetch DB existence check.\nAlready-known URLs are still fetched over HTTP,\nthen deduped only at INSERT OR IGNORE."]
```

If this summary ever disagrees with `openspec/specs/`, treat the specs as canonical.
