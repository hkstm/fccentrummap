## REMOVED Requirements

### Requirement: Extraction input uses sentence-level transcript units with timestamps
**Reason**: The transcript-based extraction flow is removed because Gemini-direct extraction is now the source of truth.
**Migration**: Use Gemini-direct extraction with article URL and YouTube URL inputs.

### Requirement: Prompt requires Dutch extraction focused on Amsterdam region
**Reason**: The transcript-based extraction prompt is removed with the legacy extraction flow.
**Migration**: Use the Gemini-direct prompt that extracts presenter, spots, optional addresses, timestamps, evidence, and confidence from article/video context.

### Requirement: Response must be parseable into timestamp-refined extraction contract
**Reason**: The two-pass transcript timestamp refinement contract is removed with the legacy extraction flow.
**Migration**: Validate Gemini-direct structured responses and persist accepted timestamp values from the Gemini response.

### Requirement: Unified extract-spots stage keeps SQLite record and artifact outputs
**Reason**: The legacy `extract-spots` transcription stage is removed from the current pipeline.
**Migration**: Use `extract-spots-gemini-direct`, which persists Gemini spot mentions and writes reviewable prompt/response/parsed artifacts.

### Requirement: Expensive-stage retry policy is operator-controlled
**Reason**: This requirement applied to the removed transcript-based extraction stage.
**Migration**: Keep operator-controlled retries for Gemini-direct model calls through explicit reruns rather than automatic retry behavior.
