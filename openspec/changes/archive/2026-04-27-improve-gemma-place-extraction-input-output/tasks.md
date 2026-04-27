## 1. Persistence schema reset

- [x] 1.1 Implement destructive migration to drop/recreate extraction storage with `presenter_name` in the new schema.
- [x] 1.2 Update storage/domain structs and row mapping to the recreated schema.
- [x] 1.3 Add migration safety notes and rollback instructions for restoring prior schema backup.

## 2. Prompt input contract updates

- [x] 2.1 Update extraction prompt builder to include both cleaned article text and transcript sentence units as separately labeled sections.
- [x] 2.2 Add explicit prompt instructions: article is primary for place identification, transcript is required for evidence/timestamp alignment.
- [x] 2.3 Ensure extraction run fails before model call when sentence-level transcript units are unavailable.

## 3. Model output parsing and validation

- [x] 3.1 Extend response schema/parser to map top-level `presenter_name` from model output.
- [x] 3.2 Keep `presenter_name` nullable/empty without failing valid place+timestep extraction.
- [x] 3.3 Enforce that every accepted extracted place has transcript-backed `sentenceStartTimestamp`.

## 4. Debug artifact capture and test coverage

- [x] 4.1 Persist final rendered prompt text sent to Gemma as a debug artifact for verification runs.
- [x] 4.2 Persist raw model response text/JSON as a debug artifact before/alongside parsed output.
- [x] 4.3 Add/update deterministic fixtures and tests for dual-source prompt input plus `presenter_name` output handling.

## 5. Targeted verification run

- [x] 5.1 Run extraction only for `https://fccentrum.nl/story/de-spots-van-niels-oosthoek/` on the new schema.
- [x] 5.2 Confirm outputs include transcript-backed places and top-level `presenter_name`.
- [x] 5.3 Review saved prompt and raw model output artifacts and document verification findings.
