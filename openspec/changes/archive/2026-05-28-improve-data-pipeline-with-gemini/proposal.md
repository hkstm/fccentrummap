# Improve Data Pipeline with Gemini

## Summary

Update the Gemini-based extraction and export pipeline to include richer spot metadata, simplify frontend presenter handling, and remove the legacy transcription-based flow now that the Gemini flow is the source of truth.

## Motivation

The current pipeline still carries legacy transcription-era code and data structures that are no longer needed. Gemini extraction is now reliable enough to become the primary flow. We also want to improve downstream geocoding quality by capturing addresses directly from videos when Gemini can identify them with reasonable confidence.

## Proposed Changes

### Data export

- Include each spot's primary type display name in the exported `spots` array.
- Remove the separate exported `presenters` array.
- On the frontend, derive presenters from the `spots` array for now by extracting presenter names and de-duplicating them.
- Defer primary-type display name filtering to a later iteration.

### Gemini extraction

- Update the Gemini prompt to extract an optional address for each spot.
- Store the extracted address in the Gemini spots table as a nullable field.
- Prompt guidance should explain that transition cards often show the spot name and may include the address as a subtitle, but this is not guaranteed. Gemini should only return an address when it is reasonably confident.
- Improve presenter extraction by instructing Gemini to use the article title and video title to determine the canonical presenter name and spelling, preferring the form used in the title when it appears to represent the name the person is known by.

### Geocoding

- When geocoding a spot, include the stored address in the geocoding request when available.

### Legacy cleanup and reprocessing

- Remove all code related to the old transcription-based extraction flow.
- Clean up the database schema/data that supported the old transcription flow.
- Drop all previously extracted Gemini data.
- Re-run extraction and geocoding from scratch so all records use the updated prompt and address-aware geocoding behavior.

## Non-goals

- Do not add primary-type filtering in this change.
- Do not keep backwards compatibility with the old transcription flow.
