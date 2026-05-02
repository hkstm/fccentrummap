## Why

Current place extraction can anchor timestamps too late when the canonical place name is only spoken at the end of a discussion segment (for example, naming a venue after describing it for several sentences). This reduces playback usefulness, so we need a refinement step that shifts timestamps to the earliest logical mention context while keeping transcript-backed evidence.

## What Changes

- Add a second extraction/refinement pass that runs after the current place+timestamp pass.
- Feed the refinement pass with transcript sentences plus first-pass place/timestamp results (no article text required in this pass).
- Require refinement to search backward from the first-pass timestamp and return the earliest logical timestamp where the same place discussion starts.
- Allow no-op refinement: if the first-pass timestamp is already the earliest logical anchor, keep it unchanged.
- Keep guardrails so refined timestamps stay transcript-backed and do not move forward in time.
- Persist and expose refined timestamps as the final timestamp output for extracted places.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `transcript-spot-extraction`: Extend extraction behavior from single-pass timestamp anchoring to two-pass refinement, with earliest-logical-discussion timestamp selection constrained by transcript evidence and explicit support for unchanged timestamps when already optimal.
- `gemma-generate-content-client`: Add a refinement prompt/response interaction for timestamp backtracking using first-pass results and transcript sentence units.
- `extraction-dry-run-artifacts`: Include refinement-pass prompt/response artifacts so timestamp adjustments can be inspected during dry runs.

## Impact

- Affected code: `scraper/internal/extraction/*`, extraction orchestration in `scraper/cmd/extract-spots-dry-run`, and related parsing/validation tests.
- Affected output contract: extraction flow continues to return place + timestamp, but timestamp provenance now includes a refinement step.
- No new external dependencies expected; existing model client is reused for an additional call per run.