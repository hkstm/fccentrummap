## ADDED Requirements

### Requirement: Frontend consumes exported static data
The frontend area under `viz/` SHALL consume generated static JSON and SHALL NOT depend on direct SQLite access.

#### Scenario: Frontend data access
- **WHEN** the frontend loads spot data
- **THEN** it SHALL read the generated JSON artifact rather than opening the SQLite database directly

### Requirement: Frontend remains host-portable
Frontend implementation choices SHALL remain portable across non-Vercel environments.

#### Scenario: Generic deployment target
- **WHEN** the frontend is deployed outside Vercel
- **THEN** its core application behavior SHALL remain supportable without Vercel-specific services

### Requirement: Business logic is framework-portable
Frontend business logic SHALL live in plain TypeScript modules rather than being embedded exclusively in host-specific framework surfaces.

#### Scenario: Reusing frontend logic
- **WHEN** frontend data shaping or domain logic is needed outside a specific route or hosting surface
- **THEN** that logic SHALL be available from plain TypeScript modules
