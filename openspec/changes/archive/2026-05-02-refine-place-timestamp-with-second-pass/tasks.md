## 1. Batched refinement contract and prompt/response plumbing

- [x] 1.1 Add extraction model types for pass-1 `originalSentenceStartTimestamp` and pass-2 `refinedSentenceStartTimestamp` per place.
- [x] 1.2 Implement a Dutch pass-2 batched refinement prompt builder that accepts transcript sentence units plus all pass-1 place/timestamp results.
- [x] 1.3 Extend Gemma function-calling config/schema to support pass-2 batched timestamp refinement output.
- [x] 1.4 Add parser/validator for pass-2 response that maps refined timestamps to pass-1 places and rejects malformed items.

## 2. Two-pass extraction orchestration and guardrails

- [x] 2.1 Update extraction orchestration to run pass 1 first, then run pass 2 as a single batched call when pass-1 places exist.
- [x] 2.2 Enforce pass-1-order-based validation: refined timestamp must be `<=` original and strictly greater than the previous place timestamp in pass-1 order.
- [x] 2.3 Implement no-op fallback per place so invalid/missing refinement keeps original timestamp unchanged.
- [x] 2.4 Ensure final extraction result exposes both original and refined timestamps, with refined timestamp used as primary downstream timestamp.

## 3. Dry-run artifacts and command integration

- [x] 3.1 Extend `extract-spots-dry-run` to emit both pass-1 and pass-2 prompt artifacts with deterministic filenames.
- [x] 3.2 Extend `extract-spots-dry-run` to emit both pass-1 and pass-2 raw response artifacts with deterministic filenames.
- [x] 3.3 Update dry-run console output to report all generated two-pass artifact paths.
- [x] 3.4 Keep dry-run non-persistence behavior unchanged (no extracted-place DB writes).

## 4. Persistence and tests

- [x] 4.1 Update extraction persistence model/schema handling to store both `originalSentenceStartTimestamp` and `refinedSentenceStartTimestamp` per place.
- [x] 4.2 Add unit tests for refinement validation rules: earlier acceptance, equal/no-op acceptance, later rejection, and previous-place-bound rejection.
- [x] 4.3 Add integration-style extraction flow test covering full two-pass batched execution with multi-place output.
- [x] 4.4 Add dry-run artifact tests asserting both pass prompts/responses are written and DB remains unchanged.
