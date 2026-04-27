# Verification findings: improve-gemma-place-extraction-input-output

Date: 2026-04-27

## Command run

```bash
cd scraper && go run ./cmd/extract-spots-dry-run \
  --db-path ../data/spots.db \
  --article-url https://fccentrum.nl/story/de-spots-van-niels-oosthoek/ \
  --reset-extraction-storage \
  --out-dir ../data
```

## Scope confirmation

- Extraction was run for exactly:
  - `https://fccentrum.nl/story/de-spots-van-niels-oosthoek/`
- Selected transcription:
  - `transcription_id=1`

## Artifacts captured

- Prompt artifact:
  - `data/transcript_extraction_1_20260427T184537Z_prompt.txt`
- Raw model response artifact:
  - `data/transcript_extraction_1_20260427T184537Z_response.json`
- Additional context artifacts:
  - `data/transcript_extraction_1_20260427T184537Z_article.txt`
  - `data/transcript_extraction_1_20260427T184537Z_transcript.json`

## Output checks

Raw response function call args contained:

- `presenter_name`: `Niels Oosthoek`
- spots:
  - `Nationale Opera & Ballet` @ `15.13`
  - `Casa del Gusto` @ `214.72`
  - `Order Tattoos` @ `433.02`

Validation checks passed:

- Every extracted spot includes `sentenceStartTimestamp`
- Top-level `presenter_name` is present and persisted

## Persistence checks

`article_spot_extractions` latest row:

- `spot_extraction_id=1`
- `article_raw_id=1`
- `transcription_id=1`
- `presenter_name='Niels Oosthoek'`
- prompt/raw/parsed payloads persisted in DB columns

## Migration safety

- Destructive reset was executed with backup preservation.
- Backup table created:
  - `article_spot_extractions_backup_20260427T184537Z`
- Rollback instructions documented in:
  - `docs/spot-extraction-storage-migration.md`
