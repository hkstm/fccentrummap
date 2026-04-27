# Spot extraction storage migration safety notes

This project now uses a destructive reset for `article_spot_extractions` when `extract-spots-dry-run` is run with:

```bash
go run ./cmd/extract-spots-dry-run --reset-extraction-storage ...
```

## What the reset does

- Renames the current table to a backup table:
  - `article_spot_extractions_backup_<UTC timestamp>`
- Recreates `article_spot_extractions` with the new schema including:
  - `presenter_name`
  - `prompt_text`
  - `raw_response_json`
  - `parsed_response_json`

## Safety guidance

- Use this only when a destructive reset is acceptable.
- Always verify the backup table name printed in logs.
- Keep the backup table until verification is complete.

## Rollback / restore instructions

If you need to restore old data after reset:

> ⚠️ This rollback is destructive for post-reset data in `article_spot_extractions`.
> Any extraction rows written after running `--reset-extraction-storage` will be lost
> if you drop the current table. Stop extraction CLI writers before rollback.

1. Open SQLite shell on your db.
2. (Safer) Rename the current post-reset table to preserve it for manual reconciliation.
3. Rename the original backup table back.

Safer example (recommended):

```sql
ALTER TABLE article_spot_extractions
  RENAME TO article_spot_extractions_postreset_backup_20260427T190000Z;
ALTER TABLE article_spot_extractions_backup_20260427T190000Z
  RENAME TO article_spot_extractions;
CREATE INDEX IF NOT EXISTS idx_article_spot_extractions_article_raw_id
  ON article_spot_extractions(article_raw_id);
```

Destructive example (drops post-reset rows):

```sql
DROP TABLE IF EXISTS article_spot_extractions;
ALTER TABLE article_spot_extractions_backup_20260427T190000Z
  RENAME TO article_spot_extractions;
CREATE INDEX IF NOT EXISTS idx_article_spot_extractions_article_raw_id
  ON article_spot_extractions(article_raw_id);
```

Then rerun the extraction command without `--reset-extraction-storage`.
