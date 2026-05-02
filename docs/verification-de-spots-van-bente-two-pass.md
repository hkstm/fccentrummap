# Verification findings: de-spots-van-bente (two-pass refinement)

Date: 2026-04-27

## Commands run

```bash
cd scraper && go run ./cmd/extract-spots-dry-run \
  --db-path ../data/spots.db \
  --article-url https://fccentrum.nl/story/de-spots-van-bente/ \
  --out-dir ../data
```

## Scope confirmation

- Extraction was run for exactly:
  - `https://fccentrum.nl/story/de-spots-van-bente/`
- Selected transcription:
  - `transcription_id=3`

## Artifacts captured

- Pass-1 prompt artifact:
  - `data/transcript_extraction_3_20260427T200440Z_pass1_prompt.txt`
- Pass-2 prompt artifact:
  - `data/transcript_extraction_3_20260427T200440Z_pass2_prompt.txt`
- Pass-1 raw response artifact:
  - `data/transcript_extraction_3_20260427T200440Z_pass1_response.json`
- Pass-2 raw response artifact:
  - `data/transcript_extraction_3_20260427T200440Z_pass2_response.json`
- Additional context artifacts:
  - `data/transcript_extraction_3_20260427T200440Z_article.txt`
  - `data/transcript_extraction_3_20260427T200440Z_transcript.json`

## Output checks

### Pass 1 (`submit_spots`)

- `presenter_name`: `Bente`
- spots:
  - `Kapsalon in de Jordaan` @ `19.86`
  - `Banketbakkerij Arnold Cornelis` @ `194.02`
  - `Dream Unit` @ `424.28`
  - `Fugazzi` @ `430.22`

### Pass 2 (`submit_refined_spots`)

Returned refinements:
- `Kapsalon in de Jordaan` → `13.8` (earlier)
- `Banketbakkerij Arnold Cornelis` → `194.02` (equal/no-op)

Missing in pass-2 output (fallback to pass-1 originals):
- `Dream Unit`
- `Fugazzi`

### Final (post-validation + fallback) expectation

Based on implemented guardrails (`refined <= original`, strict previous-place bound in pass-1 order, fallback on missing/invalid):

- `Kapsalon in de Jordaan`: original `19.86`, refined `13.8`, final primary timestamp `13.8`
- `Banketbakkerij Arnold Cornelis`: original `194.02`, refined `194.02`, final primary timestamp `194.02`
- `Dream Unit`: original `424.28`, refined missing (fallback), final primary timestamp `424.28`
- `Fugazzi`: original `430.22`, refined missing (fallback), final primary timestamp `430.22`

## Persistence checks

Dry-run should not write extraction rows. Verified with:

```sql
SELECT spot_extraction_id, article_raw_id, transcription_id, presenter_name, created_at
FROM article_spot_extractions
WHERE article_raw_id = 2
ORDER BY spot_extraction_id DESC
LIMIT 5;
```

Observed latest row remains unchanged:
- `spot_extraction_id=2`
- `article_raw_id=2`
- `transcription_id=3`
- `presenter_name='Bente'`
- `created_at='2026-04-27 18:53:31'`

This confirms the current two-pass dry-run execution generated artifacts but did **not** persist new DB extraction rows.
