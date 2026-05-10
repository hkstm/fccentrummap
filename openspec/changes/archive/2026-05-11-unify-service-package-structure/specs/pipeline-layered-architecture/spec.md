## MODIFIED Requirements

### Requirement: Stage-first package organization SHALL own contracts and orchestration while capability packages own reusable domain logic
The codebase SHALL organize pipeline internals by stage-first packages where each stage package owns stage DTOs, service orchestration, and ports. Reusable domain logic shared across stages SHALL live in explicitly named capability packages, while cross-stage generic utilities SHALL remain in dedicated shared packages.

#### Scenario: Locating stage implementation components
- **WHEN** maintainers inspect a stage package
- **THEN** they SHALL find stage-owned contracts and orchestration code in that package
- **AND** reusable non-stage-specific domain logic SHALL be located in explicit capability packages
- **AND** cross-stage shared primitives SHALL be located in `pipeline/common`

## ADDED Requirements

### Requirement: Stage services SHALL consume capability services through explicit boundaries
Stage orchestration in `pipeline/<stage>` SHALL depend on explicit interfaces or package-level APIs from capability packages rather than importing deprecated catch-all service packages.

#### Scenario: Wiring stage orchestration to reusable logic
- **WHEN** a stage requires reusable functionality such as content fetching, geocoding, transcription, or audio handling
- **THEN** the stage service SHALL use explicit capability package boundaries for that functionality
- **AND** it SHALL NOT introduce new dependencies on deprecated catch-all service package locations
