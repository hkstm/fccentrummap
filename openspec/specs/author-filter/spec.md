# author-filter Specification

## Purpose
Define presenter-based filtering behavior so users can reliably control which FC Centrum spots are visible on the map.
## Requirements
### Requirement: Multi-presenter filtering
The system SHALL allow users to filter visible spots by presenter/author.

#### Scenario: Toggle presenter visibility
- **WHEN** a presenter is deselected in the filter UI
- **THEN** all spots for that presenter SHALL be hidden from the map

### Requirement: Default all-selected filter state
The system SHALL default all presenters to selected when data loads.

#### Scenario: Initial filter state
- **WHEN** the page first loads valid spot data
- **THEN** all presenter filters SHALL start enabled and all spots SHALL be visible

### Requirement: Bulk filter controls
The system SHALL provide select-all and deselect-all controls.

#### Scenario: Select all presenters
- **WHEN** the user triggers select-all
- **THEN** all presenter filters SHALL be enabled and all corresponding spots SHALL become visible

