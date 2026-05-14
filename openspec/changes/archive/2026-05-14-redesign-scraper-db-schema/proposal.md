## Why

The current scraper persistence model allows a stage to write multiple tables, which makes stage ownership blurry, retry behavior harder to reason about, and failures harder to isolate. We are redesigning the schema from scratch now, so this is the right moment to enforce a strict one-stage/one-write-table contract without migration constraints.

## What Changes

- **BREAKING** Rebuild scraper SQLite schema from scratch and drop compatibility with the current DB layout.
- Define a stage ownership model where each pipeline stage may read many tables but writes to exactly one table.
- Split multi-write responsibilities into dedicated stages (for example: extraction, resolution, and linking steps) so writes are isolated per stage.
- Update stage contracts and CLI workflow to align command responsibilities with table ownership.
- Add relational constraints (primary keys, foreign keys, unique keys) to preserve idempotency and normalized SQL structure without JSON payload storage.

## Capabilities

### New Capabilities
- `stage-table-write-ownership`: Define and enforce the rule that each scraper stage writes to exactly one SQL table while allowing multi-table reads.
- `scraper-relational-schema-v2`: Introduce a new normalized relational schema for scraper pipeline outputs and downstream linking/export preparation.

### Modified Capabilities
- `sqlite-storage`: Replace existing scraper table model with the new schema and ownership boundaries.
- `unified-scrape-cli`: Update stage expectations and command flow to match one-table write ownership.
- `pipeline-layered-architecture`: Adjust stage boundaries and adapters to reflect single-table write responsibilities.

## Impact

- Affected code: `scraper/internal/repository/*`, stage sqlite adapters under `scraper/internal/pipeline/*`, and related domain services in `scraper/internal/*`.
- Affected behavior: stage execution order and persistence side effects per command.
- Data compatibility: existing SQLite DBs are intentionally invalidated; fresh initialization is required.
- APIs/dependencies: no new third-party dependencies expected; SQLite usage remains via `modernc.org/sqlite`.
