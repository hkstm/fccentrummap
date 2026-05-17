## 1. Stable spot identity

- [x] 1.1 Add a stable `spotId` to repository export DTOs and generated spot JSON without deriving it from corrected display fields
- [x] 1.2 Update frontend spot types and share-state logic to prefer stable `spotId` for spot-specific links
- [x] 1.3 Add tests proving `spotId` remains stable when exported name, place ID, coordinates, or timestamp are corrected

## 2. Correction storage

- [x] 2.1 Add idempotent SQLite schema initialization for spot correction storage keyed by stable `spotId`
- [x] 2.2 Implement repository operations to load source/effective spot values for correction prompts
- [x] 2.3 Implement repository upsert/retrieval for nullable correction fields without mutating source extraction or geocode rows
- [x] 2.4 Add repository tests for fresh schema, existing-schema migration, insert, update, and source-data preservation

## 3. Corrected place and timestamp handling

- [x] 3.1 Add a Google place-id coordinate lookup interface/client for resolving corrected `placeId` values to latitude/longitude
- [x] 3.2 Integrate place input resolution (place ID, Google Maps URL, or address) into correction saving and preserve the previous correction on lookup failure
- [x] 3.3 Implement timestamped YouTube URL parsing for correction input (`youtube.com` and `youtu.be` only) and YouTube link rewriting
- [x] 3.4 Add unit tests for place lookup success/failure, accepted timestamped YouTube URL forms, and rejected raw/non-YouTube/untimestamped timestamp inputs

## 4. Maintainer correction CLI

- [x] 4.1 Add an interactive correction CLI command that accepts a stable spot identifier or map URL containing `spot=`, plus a hide flag, without non-interactive edit flags in the first implementation
- [x] 4.2 Display current effective values and prompt for updated name, place ID, and timestamped YouTube URL with blank input preserving existing values
- [x] 4.3 Return actionable errors for unknown spot IDs, invalid or untimestamped YouTube URLs, missing Google credentials, and unresolved place IDs
- [x] 4.4 Add command-level tests using mocked repository/place lookup behavior

## 5. Export correction application

- [x] 5.1 Apply stored corrections during export so `spotName`, `placeId`, coordinates, and `youtubeLink` reflect overrides, and hidden spots are omitted
- [x] 5.2 Fail export with actionable spot-id-specific errors when correction data is invalid or incomplete
- [x] 5.3 Keep uncorrected spots exporting unchanged apart from the new stable `spotId` field
- [x] 5.4 Add export service/adapter tests for corrected, partially corrected, and invalid correction scenarios

## 6. Frontend spot identity state

- [x] 6.1 Update share-state parsing/building to use `spotId` while preserving compatibility with existing supported query behavior where practical
- [x] 6.2 Add frontend tests for opening a shared corrected spot by stable identifier

## 7. Validation

- [x] 7.1 Run Go tests for scraper/repository/export/CLI packages
- [x] 7.2 Run frontend tests and typecheck
- [x] 7.3 Regenerate static JSON only if correction or `spotId` export changes require fixture/data updates
