## Context

The project already stores article-linked audio and transcription responses in SQLite, and includes CLI-based workflows for transcription and artifact export. The new change adds an extraction stage that uses timestamped transcript sentences as model input and asks Gemma to identify 2-7 locations introduced as "De spots van" in the video context.

This work is intentionally staged: first a dry-run flow that reads one transcript from DB, writes the generated Dutch prompt to disk, calls the Google Generative Language `generateContent` endpoint, and writes raw model output to disk for review. Automatic DB persistence of extracted spots is deferred until output quality is validated.

Constraints:
- Implementation in Go, aligned with existing CLI/repository patterns.
- Source transcript data comes from existing transcription storage.
- Canonical extraction input is sentence-level transcript segments with timestamps only (at minimum `text` + `start`, optionally `end`).
- Full-text transcript blobs and word-level token timings are explicitly out of scope for prompt input.
- Model contract should remain compatible with `/v1beta/models/{model}:generateContent` style calls.
- Output must be structured enough for later parsing while still allowing raw JSON blob storage.

## Goals / Non-Goals

**Goals:**
- Provide a reproducible CLI dry-run command for transcript-to-spot extraction.
- Generate a Dutch prompt that explicitly asks for sentence-level evidence with timestamps.
- Enforce a structured response shape that can be consumed later (2-7 extracted place candidates).
- Persist three inspectable artifacts locally: exported transcript, final prompt text, raw API response.
- Keep model/client/config surface flexible for Gemma model variants (including `gemma-4-31b-it` target).

**Non-Goals:**
- Writing extracted spot results back into normalized DB tables.
- Building final geocoding or map rendering from extracted spots.
- Designing a perfect final prompt; prompt iteration is expected after manual review.
- Broad batch processing of all transcripts in this phase.

## Decisions

1. **Add a dedicated CLI dry-run subcommand rather than embedding behavior into existing transcription commands.**
   - **Why:** Keeps extraction concerns separate from transcription acquisition/storage and makes testing easy.
   - **Alternative considered:** Reusing existing transcription CLI flags; rejected because mixed responsibilities reduce clarity and testability.

2. **Read exactly one transcript record per run (explicit ID preferred, latest fallback optional).**
   - **Why:** Matches the validation-first scope and minimizes moving parts.
   - **Alternative considered:** Batch extraction; rejected for now due to higher error-handling and review complexity.

3. **Build prompts from sentence-level transcript units and include timestamps directly in prompt context. Sentence-level units are the canonical source format for extraction.**
   - **Why:** User requirement explicitly asks for sentence-level extraction with timestamp traceability.
   - **Alternative considered:** Passing full text blob only; rejected because it weakens evidence mapping and may produce untraceable outputs.

4. **Explicitly exclude full-text transcript blobs and word-level timings from model prompt input.**
   - **Why:** These granularities either lose sentence-to-evidence structure (full text) or add unnecessary noise/volume (word tokens) for the current extraction goal.
   - **Alternative considered:** Including all transcript representations in prompt context; rejected due to prompt bloat and reduced extraction precision.

5. **Require strict JSON output schema from the model, but store raw response payload unchanged as a dry-run artifact.**
   - **Why:** Structured response is needed for downstream parsing, while raw payload preserves debugging fidelity.
   - **Alternative considered:** Parse-and-store only normalized output; rejected because early prompt/model tuning benefits from full raw response.

6. **Use a small adapter client for Google Generative Language `generateContent`.**
   - **Why:** Encapsulates endpoint/model/API-key configuration and request/response handling.
   - **Alternative considered:** Inline `http.Client` calls in command code; rejected for maintainability and testability.

7. **Standardize local artifact file naming under `data/` with deterministic prefixes and run timestamps.**
   - **Why:** Makes manual inspection and reruns straightforward without accidental overwrites.
   - **Alternative considered:** Temporary files only; rejected because persistent inspectable outputs are a primary goal.

8. **Treat 2-7 extracted candidates as a model target, not a hard runtime validation gate.**
   - **Why:** The prompt should ask the model to find 2-7 places, but command success should not depend on receiving that exact count.
   - **Alternative considered:** Failing command execution when output count is outside 2-7; rejected because it creates false negatives during prompt/model iteration.
   - **Validation boundary:** The command fails only when the model response cannot be parsed/validated as the expected structured format.

9. **Use a minimal extraction contract for downstream ingestion: `place` and `sentenceStartTimestamp` only.**
   - **Why:** These two fields are sufficient for current storage and follow-up processing needs.
   - **Alternative considered:** Requiring extra fields (e.g., confidence, quote span, sentence index); rejected as unnecessary at this phase.

10. **Include geographic disambiguation guidance in the Dutch prompt: extracted places must be in Amsterdam or nearby surroundings.**
   - **Why:** Location mentions can be ambiguous; geographic scope improves precision for downstream map/use-case alignment.
   - **Alternative considered:** No region hint in prompt; rejected because it increases off-target place extraction.



## Risks / Trade-offs

- **[Risk] Prompt ambiguity yields non-JSON or low-quality extractions** → **Mitigation:** Include explicit JSON schema instructions in Dutch, add response validation, keep raw response for diagnosis.
- **[Risk] Transcript sentence/timestamp data is inconsistent across historical rows** → **Mitigation:** Define minimal accepted sentence object shape and fail fast with actionable errors when missing.
- **[Risk] Model version naming differences (`gemma-4-31b-it` vs available hosted variants) cause API failures** → **Mitigation:** Make model name configurable with sensible defaults and clear CLI error messages.
- **[Trade-off] Raw response storage without immediate DB integration delays automation** → **Mitigation:** This is intentional for quality gating; DB write-back will be introduced in a follow-up change.

## Migration Plan

1. Introduce CLI command + model client + prompt builder + artifact writer behind new command path.
2. Run dry-run against one known transcript ID and inspect generated files.
3. Iterate prompt contract until stable extraction quality is reached.
4. Once approved, follow-up change adds DB persistence and downstream processing.

Rollback strategy:
- If unstable, remove/disable the new dry-run command path without affecting existing transcription and export workflows.

## Open Questions

- None at this stage.