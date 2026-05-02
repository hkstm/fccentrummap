## Context

The current extraction flow produces place candidates with transcript timestamps in a single model pass. In edge cases where a place is discussed before it is explicitly named, the returned timestamp can be anchored to the end of the discussion instead of the start. We want a low-friction improvement that keeps existing extraction behavior, adds minimal new complexity, and improves timestamp usefulness in playback/search contexts.

Constraints:
- Keep transcript-backed evidence requirements for all final timestamps.
- Avoid adding new dependencies or major data-model changes.
- Preserve current behavior when no better earlier anchor exists.
- Keep dry-run debuggability by exporting intermediate artifacts.

## Goals / Non-Goals

**Goals:**
- Introduce a second-pass timestamp refinement step that runs after first-pass place extraction.
- Refine each place timestamp by searching for the earliest logical point in transcript context prior to (or equal to) the first-pass timestamp.
- Support no-op outcomes where the original timestamp is already the best anchor.
- Enforce deterministic guardrails in code so refinement cannot move timestamps forward.
- Make refinement prompt/response visible in dry-run artifacts.

**Non-Goals:**
- Replacing first-pass place detection logic or switching to full section-based modeling in this change.
- Using article text in the second pass.
- Introducing confidence scoring as a required persisted field.
- Broad schema/database redesign.

## Decisions

1. **Two-pass architecture with transcript-only batched refinement pass**
   - Decision: Keep pass 1 as canonical place detection + initial timestamp; add pass 2 as a single batched call that receives all extracted places + timestamps and returns refined timestamps for all places.
   - Rationale: Minimal disruption to proven extraction path while isolating the specific late-anchor failure mode, while keeping LLM reasoning consistent across places in one transcript.
   - Alternatives considered:
     - Per-place refinement calls: simpler payloads but more call overhead and weaker global consistency.
     - Single-pass prompt-only tweak: lower implementation effort, but less reliable and harder to debug systematically.
     - Full section detection + mapping pipeline: likely strongest long-term quality, but larger scope than needed now.

2. **Refinement search window is ordered by neighboring place timestamps**
   - Decision: For each place in sorted timestamp order, pass 2 may search backward from that place’s original timestamp down to a dynamic lower bound defined by the previous place’s pass-1 original timestamp. For the first place, the lower bound is the start of the audio.
   - Rationale: This preserves chronological ordering between places while allowing useful backtracking per place without crossing the previous pass-1 anchor.
   - Alternatives considered:
     - Fixed bounded window (e.g., N seconds): rejected because it can hide valid earlier anchors.

3. **No-op refinement is first-class behavior**
   - Decision: If pass 2 cannot find an earlier logical anchor with transcript evidence, keep the original timestamp unchanged.
   - Rationale: Avoids forced regressions and preserves correctness when first-pass already found earliest point.
   - Alternatives considered:
     - Always require earlier timestamp: would create false shifts and lower trust.

4. **Deterministic post-validation in application layer**
   - Decision: Validate pass-2 output per place before accepting: non-empty place mapping, timestamp present, timestamp not later than the place’s own pass-1 original timestamp, and timestamp strictly greater than the previous place’s pass-1 original timestamp (if a previous place exists).
   - Rationale: LLM output can be inconsistent; code-level guardrails keep behavior stable and keep place timestamps strictly ordered against pass-1 anchors.
   - Alternatives considered:
     - Trust model output fully: simpler but unsafe.

5. **Dry-run artifacts include both refinement prompt and raw response**
   - Decision: Extend `extract-spots-dry-run` outputs with second-pass prompt/response files alongside existing artifacts.
   - Rationale: Faster iteration/debugging without DB writes; mirrors current observability pattern.
   - Alternatives considered:
     - No artifact output for pass 2: harder to inspect and tune prompts.

6. **Persist both original and refined timestamps for auditability**
   - Decision: Store both `originalSentenceStartTimestamp` (pass 1) and `refinedSentenceStartTimestamp` (pass 2) per extracted place; use the refined timestamp as the primary timestamp for downstream consumption.
   - Rationale: Keeps provenance and debugging visibility without changing downstream semantics.
   - Alternatives considered:
     - Persist only refined timestamp: simpler but loses traceability.

## Risks / Trade-offs

- **[Risk] Additional model call increases runtime and cost** → Mitigation: prioritize timestamp quality over strict payload limits; allow full relevant transcript context for pass 2 when beneficial, and optionally skip refinement when candidate set is empty.
- **[Risk] Over-backtracking to loosely related earlier chatter** → Mitigation: enforce per-place dynamic lower bound from previous place pass-1 timestamp (or audio start for first place) plus transcript evidence requirement.
- **[Risk] Place-name mismatch between pass 1 and pass 2 mapping** → Mitigation: pass place labels from pass 1 verbatim and validate per-item mapping before applying.
- **[Trade-off] Better timestamp quality vs. higher pipeline complexity** → Mitigation: prioritize correctness and maintainability first, even if CLI/output contracts need to change during development; keep changes explicit and documented across extraction, dry-run artifacts, and persistence.

## Migration Plan

1. Add refinement prompt builder/config and response parsing in `scraper/internal/extraction`.
2. Update extraction orchestration to perform pass 1 then pass 2, applying guarded timestamp replacement.
3. Extend dry-run command artifact generation with refinement prompt/response outputs.
4. Add/update tests for:
   - no-op refinement,
   - earlier timestamp acceptance,
   - later timestamp rejection,
   - missing/invalid refinement fields fallback.

## Open Questions

- None.