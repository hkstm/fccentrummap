## Context

The current spot pipeline extracts places after audio acquisition and Murmel transcription. `extract-spots` builds Dutch prompts from cleaned article text plus sentence-level transcript units, calls Gemini in two function-calling passes, persists accepted mentions to SQLite, and later `geocode-spots` resolves those mentions for export.

The experiment is to evaluate whether Gemini can extract the same practical data directly from article context and the YouTube URL, reducing reliance on transcription while keeping the existing pipeline as the baseline. The project already has a Gemini client, stage-first pipeline packages, deterministic artifact practices, SQLite as canonical persistence, and Google Places geocoding/export stages.

## Goals / Non-Goals

**Goals:**

- Add an explicit experimental CLI path for Gemini-direct extraction from article context plus YouTube URL.
- Capture prompt, raw response, and parsed output artifacts for manual review and repeatable comparison.
- Produce a structured comparison report against existing transcript-based extraction results for the same article/source data.
- Keep current `extract-spots`, `geocode-spots`, corrections, and `export-data` behavior unchanged.
- Design the output shape so Gemini-direct results can later be promoted into geocoding/export if the experiment proves reliable.

**Non-Goals:**

- Do not remove Murmel transcription or the transcript-based extraction flow in this change.
- Do not make Gemini-direct results the default source for `spots.json`.
- Do not rely on unstructured natural-language model answers as the persisted comparison contract.
- Do not add frontend UI for report review in this change.

## Decisions

### Decision 1: Add a separate experimental stage instead of modifying `extract-spots`

Create a distinct stage, tentatively `extract-spots-gemini-direct`, wired through the unified scrape CLI and implemented with the same stage-first package pattern used under `scraper/internal/pipeline`.

Rationale:

- Keeps the baseline extraction path stable and comparable.
- Makes operator intent explicit; experimental model calls are not hidden inside the canonical pipeline.
- Avoids accidentally changing `spots.json` or existing DB write semantics.

Alternatives considered:

- Add a flag to `extract-spots`: simpler CLI surface, but risks mixing baseline and experiment logic.
- Replace `extract-spots`: too risky before comparison data exists.

### Decision 2: Start with artifact-first Gemini-direct output

The first implementation should persist Gemini-direct prompt, raw response, and parsed candidate JSON artifacts under a deterministic output directory, e.g. `data/gemini-direct/`, while leaving canonical extraction tables unchanged.

Rationale:

- The purpose is evaluation, not production replacement.
- Artifact-only output avoids schema churn while prompt/schema quality is still being tuned.
- Existing raw-response artifact practice already supports model debugging.

Alternatives considered:

- Add dedicated SQLite tables immediately: useful for repeatable querying, but premature if the response contract changes during experimentation.
- Write into existing `spot_mentions`: would pollute the baseline and make comparison ambiguous.

### Decision 3: Use URLs-only input and structured Gemini output with evidence fields

Gemini-direct prompts should use the article URL and YouTube URL as the primary model input for now, not cleaned article text from SQLite. The prompt should require structured output via function calling or response schema. Parsed candidates should include at least presenter, place, YouTube timestamp seconds, source/evidence text, and confidence/notes.

Rationale:

- URLs-only input tests the actual simplification hypothesis: Gemini can work from article/video context without our article-text extraction or transcription stages.
- Comparison and geocoding need stable fields, not prose.
- Evidence makes manual review possible when Gemini uses video knowledge that is not represented in our transcript artifacts.
- Confidence/notes help triage candidate-only results.

Alternatives considered:

- Include cleaned article text from SQLite: likely improves grounding, but keeps dependency on part of the current pipeline and weakens the experiment.
- Freeform prompt answers: fast to test manually, but hard to compare and unsafe to automate.
- Reuse the existing pass-1/pass-2 schema unchanged: insufficient because Gemini-direct needs evidence and may not have sentence-level transcript anchors.

### Decision 4: Add a dedicated comparison report stage

Create a separate command, tentatively `compare-spot-extractions`, that reads baseline extraction data from SQLite and Gemini-direct parsed artifacts, then writes a human-readable Markdown report for side-by-side review per article.

Rationale:

- Separates model invocation cost from cheap report iteration.
- Allows rerunning comparisons after matching/report formatting changes without calling Gemini again.
- Keeps report generation deterministic and testable.
- Produces the artifact format the operator needs for review: a Markdown document with per-article side-by-side differences.

The comparison report should present, per article:

- baseline presenter vs Gemini-direct presenter
- baseline spots and Gemini-direct spots side by side
- matched spots by normalized place name
- baseline-only spots
- Gemini-direct-only spots
- timestamp values for matched spots without applying warning thresholds
- malformed candidate notes when a Gemini-direct artifact cannot be parsed

This change should skip geocoding Gemini-direct-only candidates. Existing baseline geocoding data may be shown as contextual information when already available, but the report should not issue new Google Places requests.

Concrete Markdown skeleton:

```md
# Gemini Direct Spot Extraction Comparison

Generated: <timestamp>
Baseline source: transcript-based extraction from <db-path>
Candidate source: Gemini-direct artifacts from <artifact-dir>

## Summary

| Metric | Count |
| --- | ---: |
| Articles compared | <n> |
| Presenter matches | <n> |
| Matched spots | <n> |
| Baseline-only spots | <n> |
| Gemini-direct-only spots | <n> |
| Parse failures | <n> |

## Articles

### <article title or URL>

- Article URL: <url>
- YouTube URL: <url>

#### Presenter

| Baseline | Gemini-direct | Result |
| --- | --- | --- |
| <name> | <name> | match/mismatch/missing |

#### Spots side by side

| Baseline spot | Baseline timestamp | Gemini-direct spot | Gemini timestamp | Match type | Evidence / notes |
| --- | ---: | --- | ---: | --- | --- |
| <place> | <seconds> | <place> | <seconds> | matched | <evidence> |
| <place> | <seconds> | — | — | baseline-only | — |
| — | — | <place> | <seconds> | gemini-only | <evidence> |

#### Raw notes

- <parse issue, ambiguity, or manual-review note>
```

Alternatives considered:

- Compare inside the extraction stage: convenient, but couples expensive model calls to report generation.
- Compare and geocode Gemini-direct-only candidates immediately: useful for stronger identity matching, but adds cost, side effects, and another variable before we know whether the direct extraction is reliable.
- Compare only by raw names: easiest, but normalized name matching gives a small amount of resilience without introducing geocoding.

### Decision 5: Reuse existing Gemini credential/model configuration

The direct stage should use the existing Gemini API key resolution and model flag patterns, with a plain/experimental content config rather than the current transcript extraction function declarations.

Rationale:

- Operators already know how to configure Gemini for this project.
- Existing `genai.Client` captures raw response bytes and endpoint overrides, which are useful for tests and artifacts.
- The response contract is different enough to warrant a new config builder rather than bending the transcript extraction config.

Alternatives considered:

- Direct REST calls: more control, but duplicates the existing client behavior.
- New dependency/client abstraction: unnecessary unless Gemini-direct needs capabilities unavailable in the current SDK wrapper.

## Risks / Trade-offs

- **Gemini may hallucinate places or timestamps** → Require structured evidence fields, keep baseline comparison, and do not export Gemini-direct by default.
- **Model knowledge may vary by model/version/date** → Persist prompt, raw response, parsed artifacts, model identifier, and run metadata.
- **YouTube video access through Gemini may be inconsistent** → Treat failures as explicit experimental-stage failures and preserve diagnostics.
- **Name-only matching may create false mismatches or false matches** → Use normalized name matching and keep ambiguous matches visible in the Markdown report; defer new geocoding until a follow-up change.
- **Artifact-only storage limits aggregate querying** → Accept this during the experiment; add SQLite tables later only if the contract stabilizes.
- **Extra CLI stages increase workflow complexity** → Prefix/document them as experimental and keep canonical `make scrape`/`make export` behavior unchanged.

## Migration Plan

1. Add experimental CLI commands and artifact contracts without changing existing stage defaults.
2. Run Gemini-direct extraction for a small set of articles that already have baseline transcript extraction results.
3. Generate comparison reports and manually review candidate-only/baseline-only/timestamp-delta cases.
4. If results are strong, propose a follow-up change for optional geocoding/export from Gemini-direct results or for replacing transcription in selected workflows.

Rollback is straightforward: remove/ignore the experimental artifact directory and do not invoke the new commands. Existing DB data and `spots.json` generation remain unaffected.

## Open Questions

