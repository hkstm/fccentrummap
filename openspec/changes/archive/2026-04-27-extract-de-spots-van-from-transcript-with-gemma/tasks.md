## 1. CLI command scaffolding and configuration

- [x] 1.1 Add a dedicated transcript-to-spot dry-run CLI command entrypoint under the existing Go CLI.
- [x] 1.2 Add command flags for transcript selection (explicit ID and optional latest fallback behavior).
- [x] 1.3 Add configuration loading/validation for Google Generative Language API key and Gemma model identifier.
- [x] 1.4 Return actionable command errors when required config is missing.

## 2. Transcript selection and sentence-level input validation

- [x] 2.1 Implement repository query to load one transcription record for dry-run export.
- [x] 2.2 Parse transcript payload into sentence-level units with required `text` and `start` timestamp fields.
- [x] 2.3 Enforce validation that full-text-only and word-level-only data are rejected for extraction input.
- [x] 2.4 Add clear failure messages when sentence-level timestamped input is unavailable.

## 3. Dutch prompt builder and extraction contract

- [x] 3.1 Implement Dutch prompt template that asks for "De spots van" extraction from provided sentence/timestamp units.
- [x] 3.2 Include Amsterdam/nearby-region disambiguation guidance in the prompt.
- [x] 3.3 Instruct model to target 2-7 candidates while documenting this as non-fatal if response count differs.
- [x] 3.4 Define and validate minimal parsed extraction schema with required fields: `place` and `sentenceStartTimestamp`.

## 4. Gemma generateContent client integration

- [x] 4.1 Implement a dedicated client adapter for `/v1beta/models/{model}:generateContent` requests.
- [x] 4.2 Build request payload using `contents[].parts[].text` with the composed Dutch prompt.
- [x] 4.3 Implement HTTP response handling for success and non-2xx failures with diagnostic details.
- [x] 4.4 Add parser/validator wiring that fails only on unparseable/invalid structured responses.

## 5. Dry-run artifact outputs and verification

- [x] 5.1 Write exported transcript artifact to `data/` with deterministic filename.
- [x] 5.2 Write full composed prompt text artifact to `data/`.
- [x] 5.3 Write raw model response payload artifact to `data/` unchanged.
- [x] 5.4 Ensure dry-run command does not write extracted results back to DB in this phase.

## 6. Testing and manual acceptance loop

- [x] 6.1 Add unit tests for sentence-level extraction input validation and rejection paths.
- [x] 6.2 Add unit tests for prompt builder (Dutch language instructions, Amsterdam scope, minimal output contract).
- [x] 6.3 Add tests for client request formatting and non-2xx error surfacing.
- [x] 6.4 Run one end-to-end dry-run against a real transcript and verify transcript/prompt/response artifacts are inspectable for approval.