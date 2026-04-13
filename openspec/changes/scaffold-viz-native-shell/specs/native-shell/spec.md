## ADDED Requirements

### Requirement: Frontend workspace may be wrapped by a Tauri native shell
The frontend workspace under `viz/` SHALL support a scaffolded Tauri wrapper for future native packaging.

#### Scenario: Native shell scaffold exists
- **WHEN** a contributor inspects `viz/`
- **THEN** they SHALL find placeholder Tauri configuration and native-shell project files alongside the frontend workspace

### Requirement: Native shell preserves the frontend data boundary
The Tauri shell SHALL host the frontend without bypassing the exported JSON boundary.

#### Scenario: Native shell data access
- **WHEN** the frontend runs inside the native shell
- **THEN** it SHALL continue to consume exported app data from the frontend asset layer rather than reading SQLite directly

### Requirement: Native shell scaffold is non-product-functional by default
The initial native shell scaffold SHALL provide tooling and project structure without claiming to implement end-user product features.

#### Scenario: Initial scaffold state
- **WHEN** the scaffold change is complete
- **THEN** the native shell project SHALL contain placeholder config, metadata, and starter files only
