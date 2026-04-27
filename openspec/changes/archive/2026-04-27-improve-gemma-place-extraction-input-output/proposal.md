## Why

The current Gemma extraction flow appears to rely primarily on transcript text, which can miss or misclassify place mentions that are clearer in cleaned article content. We also only extract places today, but downstream usage now needs attribution of the single primary presenter for the extracted spots.

## What Changes

- Extend Gemma prompt input to include a cleaned article representation alongside transcript data.
- Define the cleaned article as the primary source of truth for place identification when place mentions are present.
- Keep transcript timing alignment as a first-class requirement so identified places are still mapped to when they are mentioned in the video.
- Extend model output schema to include extracted places plus one top-level primary presenter/person field.
- Update validation and storage contracts so presenter attribution is persisted and consumable.
- Apply a destructive persistence migration (drop/recreate extraction table/model) to adopt the new schema cleanly.

## Capabilities

### New Capabilities
- `spot-mention-attribution`: Capture and return the single primary presenter/person associated with the extracted spots.

### Modified Capabilities
- `transcript-spot-extraction`: Prioritize cleaned article content for place identification while preserving transcript-based temporal anchoring.
- `gemma-generate-content-client`: Expand prompt construction and structured output parsing to support cleaned article input and single-presenter-attributed output.

## Impact

- Affected code in prompt-building and response parsing for Gemma extraction.
- Updates to extraction/domain models and SQLite persistence schema, including destructive table/model reset for the new attribution contract.
- Potential updates to any API/CLI output contracts and tests that currently assume places-only extraction.
