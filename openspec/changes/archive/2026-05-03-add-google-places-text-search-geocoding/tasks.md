## 1. Google Places text-search geocoder function

- [x] 1.1 Define/confirm geocoder function signature for single place-name input and lat/lng output with error handling
- [x] 1.2 Implement request builder for Google Places Text Search including mandatory `locationRestriction.rectangle` bounds (low `52.274525,4.711585`; high `52.461764,5.073559`) and no `locationBias`
- [x] 1.3 Implement response parsing that selects the first valid result and extracts latitude/longitude
- [x] 1.4 Implement deterministic error handling for empty query, missing API key/config, no results in restriction, and upstream HTTP/API failures

## 2. CLI debug wrapper

- [x] 2.1 Add minimal CLI command to accept place-name input and invoke the geocoder function
- [x] 2.2 Implement default JSON output containing query, place name, place ID, and stable Google Maps URL on success
- [x] 2.3 Implement JSON-formatted error output and non-zero exit behavior on failure paths

## 3. Tests and validation

- [x] 3.1 Add unit tests for request payload construction to verify rectangle restriction values and absence of location bias
- [x] 3.2 Add unit tests for response parsing and first-result selection
- [x] 3.3 Add unit tests for deterministic error paths (missing key, empty query, no results, malformed/upstream error responses)
- [x] 3.4 Add CLI-level tests (or command execution checks) for success JSON output and failure exit behavior

## 4. Documentation and developer usage

- [x] 4.1 Document CLI usage, required environment variables, and example JSON output in README/development docs
- [x] 4.2 Document expected behavior for restriction-bounded no-result cases and troubleshooting guidance
