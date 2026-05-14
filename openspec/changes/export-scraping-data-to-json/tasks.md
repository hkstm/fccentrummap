## 1. Export contract and data mapping

- [x] 1.1 Identify SQLite source tables/joins needed to build `spots` (`placeId`, `spotName`, `presenterName`, `youtubeLink`) and `presenters`.
- [x] 1.2 Implement curated read-model query code to return export-ready structs without exposing raw table schema.
- [x] 1.3 Ensure presenter names are mapped as-is from DB values and not normalized/transformed.

## 2. JSON export writer

- [x] 2.1 Define Go export payload structs for top-level `spots` and `presenters` collections.
- [x] 2.2 Implement deterministic ordering for exported arrays (spots by `placeId`, presenters by `presenterName`).
- [x] 2.3 Implement JSON file writing to configured path with valid output for empty datasets (`spots: []`, `presenters: []`).

## 3. CLI integration

- [x] 3.1 Add CLI option to enable JSON export as an optional operation.
- [x] 3.2 Add CLI/config option for export output path.
- [x] 3.3 Wire CLI invocation to execute export after data is available in SQLite.

## 4. Validation and tests

- [x] 4.1 Add tests verifying exported schema contains required fields for spot entries.
- [x] 4.2 Add tests verifying deterministic output ordering across repeated runs with unchanged source data.
- [x] 4.3 Add tests for empty/partial dataset behavior to ensure valid JSON is always emitted.
- [x] 4.4 Add a smoke test/manual verification step for CLI export generation to a custom output path.
