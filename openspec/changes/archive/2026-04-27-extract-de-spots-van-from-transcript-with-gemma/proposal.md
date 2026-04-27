## Why

We need a repeatable way to extract "De spots van" place mentions from timestamped transcripts so we can turn long video transcriptions into structured location candidates. Doing this now enables a fast validation loop (single transcript in, model output to disk) before wiring full database write-back.

## What Changes

- Add a new extraction flow that reads one transcript with sentence-level timestamps from SQLite and builds a Dutch prompt for Gemma (`gemma-4-31b-it` target model variant).
- Add a model invocation step compatible with the Google Generative Language `generateContent` API style.
- Require structured output containing 2-7 extracted place candidates tied to transcript sentence timestamps, suitable for later DB ingestion (raw JSON blob acceptable initially).
- Add a local test path that:
  - exports one transcript from DB,
  - writes the final Dutch prompt to a text file,
  - calls the model endpoint,
  - writes raw model response to disk for manual inspection.
- Defer automatic DB persistence of extraction results until output quality is approved.

## Capabilities

### New Capabilities
- `transcript-spot-extraction`: Extract 2-7 "De spots van" place candidates from timestamped transcript sentences via a Dutch prompt and structured model output.
- `gemma-generate-content-client`: Call Google Generative Language `generateContent` for Gemma with project-configurable model and request payload.
- `extraction-dry-run-artifacts`: Support local dry-run artifacts (exported transcript, composed prompt text, raw model response file) for inspection and iteration.

### Modified Capabilities
- `audio-transcription-cli`: Extend CLI behavior with a command/subcommand path that runs transcript-to-extraction dry runs from stored transcription data.

## Impact

- Affected code: Go CLI command layer, prompt builder, model client, repository read path for transcript selection, and file output utilities.
- External API: Google Generative Language API (`/v1beta/models/*:generateContent`) for Gemma calls.
- Data/contracts: Introduces a structured extraction JSON contract (raw blob persisted later) that downstream parsing can consume.
- Dependencies/config: Requires API key + model configuration for Google endpoint access; no immediate schema migration required for first dry-run milestone.