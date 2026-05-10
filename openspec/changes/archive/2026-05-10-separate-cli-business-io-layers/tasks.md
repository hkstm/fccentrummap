## 1. Create layered pipeline module scaffold

- [x] 1.1 Create `scraper/internal/pipeline/common` with shared error/contract/artifact helper primitives used across stages.
- [x] 1.2 Create stage-first package directories for all parity-critical stages (`collectarticleurls`, `fetcharticles`, `acquireaudio`, `transcribeaudio`, `extractspots`, `geocodespots`, `exportdata`).
- [x] 1.3 Add stage-prefixed file scaffolds per package (`<stage>_dto.go`, `<stage>_ports.go`, `<stage>_service.go`, `<stage>_sqlite_adapter.go`, `<stage>_file_adapter.go`).

## 2. Define stage contracts and service boundaries

- [x] 2.1 Define typed request/response DTOs for each parity-critical stage, including deterministic artifact identity fields where needed.
- [x] 2.2 Define narrow per-stage port interfaces that isolate service logic from concrete repository/file operations.
- [x] 2.3 Implement service-level orchestration for each stage so business logic depends only on stage ports.

## 3. Implement SQLite adapters behind stage ports

- [x] 3.1 Implement SQLite adapters for `collectarticleurls`, `fetcharticles`, and `acquireaudio` by mapping existing repository-backed flows to new ports.
- [x] 3.2 Implement SQLite adapters for `transcribeaudio` and `extractspots`, preserving current persistence behavior and error semantics.
- [x] 3.3 Implement SQLite adapters for `geocodespots` and `exportdata`, preserving current support constraints (including sqlite-mode limitations where specified).

## 4. Implement file adapters with typed artifact contracts

- [x] 4.1 Replace passthrough behavior with typed file adapter input parsing and validation for all seven parity-critical stages.
- [x] 4.2 Implement deterministic stage output artifact construction per stage contract without schema-version metadata fields.
- [x] 4.3 Ensure file adapter payloads remain human-inspectable for ad-hoc debugging while still enforcing required contract fields.

## 5. Refactor CLI to delegate to stage services

- [x] 5.1 Refactor `scraper/cmd/scrape/main.go` so each stage command performs CLI concerns only (flag parsing, mode validation, service invocation, user-facing errors).
- [x] 5.2 Route `--io sqlite` and `--io file` mode selection to the corresponding stage adapter wiring.
- [x] 5.3 Remove obsolete mixed-layer command wiring paths once new service/adapters are active for all parity-critical stages.

## 6. Validate parity and behavior contracts

- [x] 6.1 Add contract parity tests executing representative scenarios against both SQLite and file adapters for all seven parity-critical stages.
- [x] 6.2 Add/update stage-level tests to verify unsupported mode handling and actionable non-zero failures remain intact.
- [x] 6.3 Run full scraper test suite and targeted smoke checks to confirm no regressions in unified CLI flows.

## 7. Update docs for layered architecture and stage behavior

- [x] 7.1 Update architecture/development docs to describe CLI → service → adapter layering and stage-first package layout.
- [x] 7.2 Document file-mode stage contracts and deterministic artifact expectations for all parity-critical stages.
- [x] 7.3 Document known backend differences (e.g., SQLite integrity guarantees vs file-mode debugging focus) without claiming unsupported capabilities.
