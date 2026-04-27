# Verification findings: de-spots-van-bente

Date: 2026-04-27

## Commands run

```bash
cd scraper && go run ./cmd/transcribe-audio \
  --db-path ../data/spots.db \
  --audio-source-id 2 \
  --language nl

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
- Persisted extraction row:
  - `spot_extraction_id=2`

## Artifacts captured

- Prompt artifact:
  - `data/transcript_extraction_3_20260427T185327Z_prompt.txt`
- Raw model response artifact:
  - `data/transcript_extraction_3_20260427T185327Z_response.json`
- Additional context artifacts:
  - `data/transcript_extraction_3_20260427T185327Z_article.txt`
  - `data/transcript_extraction_3_20260427T185327Z_transcript.json`

## Output checks

Raw response function call args contained:

- `presenter_name`: `Bente`
- spots:
  - `Kapsalon in de Jordaan` @ `19.86`
  - `Banketbakkerij Arnold Cornelis` @ `194.02`
  - `Dream Unit` @ `424.28`
  - `Bugazzi` @ `430.22`

Validation checks passed:

- Every extracted spot includes `sentenceStartTimestamp`
- Top-level `presenter_name` is present and persisted

## Persistence checks

`article_spot_extractions` row `spot_extraction_id=2`:

- `article_raw_id=2`
- `transcription_id=3`
- `presenter_name='Bente'`
- prompt/raw/parsed payloads persisted in DB columns
