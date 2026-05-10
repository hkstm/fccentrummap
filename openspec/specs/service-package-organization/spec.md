# service-package-organization Specification

## Purpose
Define canonical ownership boundaries for internal scraper service packages and deprecate legacy catch-all business-logic placement.

## Requirements

### Requirement: Internal service packages SHALL follow canonical ownership boundaries
The scraper internals SHALL organize service code using explicit ownership boundaries: stage orchestration and stage contracts in `scraper/internal/pipeline/<stage>`, reusable capability services in explicitly named capability packages, and only cross-stage primitives in `scraper/internal/pipeline/common`.

#### Scenario: Classifying new service code
- **WHEN** a maintainer adds new service logic
- **THEN** stage-specific orchestration and stage DTO/port code SHALL be placed in the owning `pipeline/<stage>` package
- **AND** reusable non-stage-specific business logic SHALL be placed in an explicitly named capability package
- **AND** domain-specific logic SHALL NOT be added to `pipeline/common`

### Requirement: Legacy catch-all service package usage SHALL be deprecated
The repository SHALL treat `scraper/internal/scraper` as a deprecated location for business logic and SHALL NOT introduce new business behavior there.

#### Scenario: Adding new business functionality
- **WHEN** a contributor implements new business logic after this change
- **THEN** the implementation SHALL be placed in stage-owned or explicit capability packages
- **AND** new logic SHALL NOT be added under `scraper/internal/scraper`

### Requirement: Service package migration SHALL preserve runtime behavior
Reorganization of service package locations SHALL preserve stage behavior, output contracts, and user-facing CLI semantics.

#### Scenario: Moving existing implementation files
- **WHEN** files are moved from legacy locations to canonical package locations
- **THEN** all parity-critical stages SHALL retain equivalent success/failure behavior
- **AND** generated artifacts and persisted outputs SHALL remain contract-compatible for existing flows
