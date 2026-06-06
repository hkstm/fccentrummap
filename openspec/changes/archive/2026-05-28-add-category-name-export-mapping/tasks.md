## 1. Schema and Repository Support

- [x] 1.1 Add `spot_category_mappings` table creation to SQLite schema initialization with primary type display name, category name, confidence, reason, model, and mapped timestamp fields
- [x] 1.2 Add repository query to fetch distinct trimmed non-empty `primary_type_display_name` values from Gemini-direct geocode rows
- [x] 1.3 Add repository write operation that replaces all `spot_category_mappings` rows in a single transaction
- [x] 1.4 Add repository/export query support to read mapped category names for exportable Gemini-direct spots with `Overig` fallback
- [x] 1.5 Add schema/repository tests for table creation, distinct primary type reads, full table replacement, and missing mapping fallback

## 2. Category Mapping Stage

- [x] 2.1 Create a category mapping package with fixed Dutch category constants and validation helpers
- [x] 2.2 Implement the English Gemini prompt builder that receives deduplicated primary type display names and includes the fixed category list
- [x] 2.3 Implement structured Gemini `GenerateContentConfig` and response schema for category mappings using the existing `internal/genai.Client` pattern
- [x] 2.4 Implement response parsing and validation for complete one-to-one mappings, known categories only, duplicate detection, and unexpected primary type rejection
- [x] 2.5 Implement the `map-spot-categories` pipeline service/adapter to read distinct primary type names, call Gemini, validate the response, and replace the mapping table snapshot
- [x] 2.6 Wire `map-spot-categories` into the scrape CLI with database path, Gemini API key, and model configuration consistent with existing Gemini stages
- [x] 2.7 Add unit tests for prompt generation, response validation failures, successful mapping replacement, and CLI/stage wiring

## 3. JSON Export Contract

- [x] 3.1 Update export models to replace `primaryTypeDisplayName` with `categoryName` on each exported spot
- [x] 3.2 Update export data model to include top-level `presenters: [{"presenterName": string}]` and `categories: [{"categoryName": string}]`
- [x] 3.3 Update Gemini-direct export repository logic to join or look up `spot_category_mappings` and set `categoryName` to `Overig` when no mapping exists
- [x] 3.4 Derive top-level presenters from final exported non-hidden spots and order them by most recent associated article publication timestamp descending with deterministic name tie-breaking
- [x] 3.5 Derive top-level categories from final exported non-hidden spots and order them by descending spot count, deterministic name tie-breaking, and `Overig` last
- [x] 3.6 Preserve deterministic spot ordering and valid JSON behavior for empty or partial datasets
- [x] 3.7 Add export tests covering mapped categories, missing mapping fallback, removal of raw primary type field, presenter metadata ordering, category metadata ordering, and empty export behavior

## 4. Frontend Static Data Consumption

- [x] 4.1 Update frontend TypeScript data types to require spot `categoryName` and include top-level `presenters` and `categories` metadata
- [x] 4.2 Update data loading/validation to parse presenter and category metadata while tolerating absent metadata arrays for empty datasets
- [x] 4.3 Update presenter filter derivation to use top-level presenter metadata order instead of deriving order from spots
- [x] 4.4 Add or update category filter/display behavior to use `categoryName` and top-level category metadata where applicable
- [x] 4.5 Remove frontend dependencies on `primaryTypeDisplayName`
- [x] 4.6 Add frontend tests for loading the new schema, presenter ordering, category metadata, and category display/filter behavior

## 5. Verification and Generated Data

- [x] 5.1 Run Go tests for repository, mapping stage, export-data, and CLI wiring
- [x] 5.2 Run frontend lint/typecheck/tests for the updated static data contract
- [x] 5.3 Run the category mapping stage against `data/spots.db` to populate `spot_category_mappings`
- [x] 5.4 Run `export-data` and verify generated `spots.json` contains `spots`, `presenters`, and `categories` with no `primaryTypeDisplayName` fields
- [x] 5.5 Run OpenSpec validation for `add-category-name-export-mapping`
