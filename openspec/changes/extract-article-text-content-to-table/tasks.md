## 1. Schema and model updates

- [x] 1.1 Add `article_text_extractions` table creation to repository schema init with required fields (`article_raw_id`, `extraction_mode`, `status`, `matched_count`, optional `error_message`, timestamps)
- [x] 1.2 Add `article_text_contents` table creation with foreign keys to `article_raw_id` and extraction outcome row
- [x] 1.3 Ensure schema initialization remains idempotent on existing databases and preserves existing data
- [x] 1.4 Add/adjust Go model structs for extraction outcome and extracted text content rows

## 2. Extraction logic implementation

- [x] 2.1 Implement article text extractor that reads raw HTML and extracts main content using Trafilatura
- [x] 2.2 Implement outcome decision logic: `trafilatura` match, `no_match` for insufficient text, and `error` for parser/runtime failures
- [x] 2.3 Normalize extracted text segments (trim, drop empty values) and attach source-type metadata
- [x] 2.4 Add unit tests for extractor scenarios: matched content, no-match, and parser error handling

## 3. Repository persistence flow

- [x] 3.1 Implement repository method to replace prior extraction outcome/content rows for an article in one transaction
- [x] 3.2 Persist one extraction outcome row for every processed article, including `no_match` and `error` outcomes
- [x] 3.3 Persist extracted content rows only for successful single-mode matches and link them to extraction outcome row
- [x] 3.4 Add repository tests (or integration tests) to verify atomic replace behavior and idempotent rerun semantics

## 4. Scraper pipeline integration

- [x] 4.1 Wire extraction execution into scraper processing after raw HTML is stored/available
- [x] 4.2 Integrate persistence call so extraction outcome and content are written per article
- [x] 4.3 Ensure extraction errors are treated as extraction-only failures (scraper run continues)
- [x] 4.4 Add logging for extraction outcomes to support drift detection (`matched`, `no_match`, `error`)

## 5. Validation and rollout checks

- [x] 5.1 Run representative end-to-end scraper run against sample FC Centrum pages and verify expected DB rows for each outcome type
- [x] 5.2 Verify no regressions in existing scraper/storage flows by running project test suite (`cd scraper && go test ./...`)
- [x] 5.3 Confirm docs/spec alignment by reviewing change artifacts and updating implementation notes if needed
