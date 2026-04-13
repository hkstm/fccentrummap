## Purpose

Define the canonical frontend boundary and portability constraints for future work in `viz/`.

## Requirements

### Requirement: Frontend consumes exported static data
The frontend area under `viz/` SHALL consume generated static JSON and SHALL NOT depend on direct SQLite access.

#### Scenario: Frontend data access
- **WHEN** frontend code loads spot data
- **THEN** it SHALL use `/data/spots.json` sourced from `viz/public/data/spots.json`
- **AND** it SHALL NOT open `data/spots.db` directly

### Requirement: Frontend remains host-portable
Frontend implementation choices SHALL remain portable across non-Vercel environments.

#### Scenario: Generic deployment target
- **WHEN** the frontend is deployed outside a vendor-specific platform
- **THEN** its core behavior SHALL remain supportable without Vercel-only services

### Requirement: Frontend business logic remains framework-portable
Frontend data shaping and domain logic SHALL live in plain TypeScript modules rather than being locked exclusively into host-specific framework surfaces.

#### Scenario: Reusing frontend logic
- **WHEN** frontend logic is needed across components, pages, or route handlers
- **THEN** that logic SHALL be available from plain TypeScript modules
