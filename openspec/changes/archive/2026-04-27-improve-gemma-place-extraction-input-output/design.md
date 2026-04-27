## Context

The current extraction flow focuses on transcript text for identifying spots and their mention timing. In practice, place names are often normalized and clearer in the cleaned article text, while transcript segments are still the best source for temporal anchoring to video timestamps. The current output model is places-only; the next stage of the pipeline requires attribution of the single primary presenter and stable persistence of that metadata.

This change crosses prompt construction, model output parsing/validation, and persistence contracts. It therefore needs a design pass to keep data contracts consistent and avoid brittle prompt/output coupling.

## Goals / Non-Goals

**Goals:**
- Add cleaned article content as an explicit Gemma prompt input channel.
- Treat cleaned article content as primary for place detection, while keeping transcript mention alignment for timing.
- Extend structured model output to include the single primary presenter attribution alongside extracted spots.
- Update parsing/validation/storage so attribution survives end-to-end.
- Keep tolerant behavior for records where attribution is missing or uncertain in newly processed data.

**Non-Goals:**
- Replacing transcript-based timing extraction with article-based timing.
- Introducing speaker diarization or audio-level person detection.
- Refactoring unrelated scraper/transcription components.
- Reworking map rendering/UI behavior in this change.

## Decisions

1. **Dual-source prompt contract with explicit precedence**
   - Decision: Prompt template will pass both `cleaned_article` and `transcript_segments` as separate labeled sections and explicitly instruct Gemma: article text is primary for identifying candidate places; transcript is primary for mention timestamps.
   - Rationale: Keeps source roles clear and reduces false negatives from noisy transcript spelling.
   - Alternative considered: Merge article and transcript into one blob. Rejected because source ambiguity harms extraction consistency and makes debugging difficult.

2. **Single-presenter attribution in output schema**
   - Decision: Extend extraction output shape with one top-level presenter field (using project naming conventions) that applies to all extracted spots.
   - Rationale: For this workflow, there is one person that matters, so a single attribution field is simpler and avoids redundant per-spot duplication.
   - Alternative considered: Per-spot `describedBy` fields. Rejected as unnecessary complexity for a single-presenter model.

3. **Graceful fallback and validation strategy**
   - Decision: Parser accepts missing attribution as nullable/empty with validation warnings instead of hard failure, while requiring place identity and timestamp anchors when available.
   - Rationale: Maintains pipeline robustness while gradually improving attribution quality.
   - Alternative considered: Hard-require attribution. Rejected because it would drop useful place extractions on imperfect inputs.

4. **Persistence contract reset via destructive migration**
   - Decision: Replace the existing extraction storage table/model with a new schema that includes single-presenter attribution as a first-class field.
   - Rationale: Project constraints allow wiping/rebuilding this table, and a clean schema is simpler than carrying compatibility baggage.
   - Alternative considered: Additive columns/JSON fields. Rejected to avoid transitional complexity and mixed-schema handling.

5. **Deterministic prompt/output fixtures for regression tests**
   - Decision: Add/update fixtures that include article + transcript inputs and expected attributed outputs.
   - Rationale: Prompt/schema changes are prone to regressions; fixtures make behavior auditable.
   - Alternative considered: Rely only on live model checks. Rejected due to nondeterminism and slower feedback.

## Risks / Trade-offs

- **[Risk] Article text may contain places not actually discussed in the video** → Mitigation: Require transcript evidence or timestamp alignment for final inclusion where possible; mark uncertain matches.
- **[Risk] Presenter name extraction can be missing/variant across inputs** → Mitigation: Allow nullable fallback rather than forcing a wrong name; defer normalization decisions until after observing real model outputs.
- **[Risk] Schema drift between prompt instructions and parser expectations** → Mitigation: Keep output schema versioned and backed by fixture tests.

## Migration Plan

1. Apply destructive migration: drop/recreate the extraction storage table/model with the new single-presenter schema.
2. Update prompt builder to include cleaned article + explicit source-precedence instructions.
3. Update response parser/validator to map attribution fields with graceful fallback.
4. Add regression fixtures and update tests for new output shape.
5. Re-run extraction only for `https://fccentrum.nl/story/de-spots-van-niels-oosthoek/` to verify end-to-end behavior before broader backfill.
6. Persist debug artifacts for that verification run: (a) final text prompt sent to Gemma, and (b) raw model output text/JSON response.
7. Rollback: restore previous schema from migration backup and revert prompt/parser changes if needed.

## Open Questions

- Canonical field name: `presenter_name` (use this across prompt output schema, domain model, and DB storage).
- Require explicit transcript evidence for every extracted place; do not include article-only places.
- No controlled vocabulary for presenter names initially; use raw model output and evaluate consistency from real runs before adding normalization.
