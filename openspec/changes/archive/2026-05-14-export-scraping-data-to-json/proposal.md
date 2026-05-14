## Why

Scraped data is currently persisted in SQLite tables, which works for ingestion and processing but not for static site delivery. We now need a static data dump so the website can bundle and serve the dataset directly when the client requests a page.

## What Changes

- Add a JSON export step that generates a static dump from the SQLite-backed scraped dataset.
- Define a stable export contract for static site consumption, including `spots` and `presenters` collections.
- Include required fields in `spots` records (`placeId`, `spotName`, `presenterName`, `youtubeLink`) and normalized presenter entries.
- Add CLI/config support to choose export output path and trigger generation.
- Document expected behavior for successful, empty, and partial exports.

## Capabilities

### New Capabilities
- `static-site-data-dump`: Export curated SQLite data into a static JSON bundle for website distribution.
- `scraping-data-json-export`: Generate a structured JSON artifact from scraped/storage data with a predictable schema.

### Modified Capabilities
- None.

## Impact

- Affected code: storage read/query layer, export pipeline stage, and CLI wiring for output controls.
- Affected systems: SQLite source-of-truth plus static site asset pipeline that consumes exported JSON.
- API/data contract impact: introduces a versionable JSON payload for frontend/static delivery.

Example target shape:

```json
{
  "spots": [
    {
      "placeId": "someId",
      "spotName": "Some spot",
      "presenterName": "Ray Fuego",
      "youtubeLink": "https://www.youtube.com/watch?v=some-video-id"
    }
  ],
  "presenters": [
    {
      "presenterName": "Ray Fuego"
    }
  ]
}
```
