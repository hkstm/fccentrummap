## ADDED Requirements

### Requirement: Spot category mapping stage
The system SHALL provide a category-mapping pipeline stage that regenerates a stored mapping from observed Google Places primary type display names to fixed frontend category names.

#### Scenario: Mapping stage classifies observed primary type display names
- **WHEN** the user runs the category-mapping stage against a database with Gemini-direct geocode rows
- **THEN** the stage MUST read distinct non-empty `primary_type_display_name` values from `gemini_direct_spot_google_geocodes`
- **AND** the stage MUST trim and deduplicate values before classification
- **AND** the stage MUST request structured Gemini output mapping each distinct value to one fixed category name

#### Scenario: Mapping stage replaces mapping snapshot
- **WHEN** the category-mapping stage receives a valid complete mapping response
- **THEN** the stage MUST replace the existing `spot_category_mappings` table contents in a transaction
- **AND** the final table MUST contain one row per distinct classified primary type display name
- **AND** the stage MUST NOT incrementally update or upsert individual mapping rows outside the complete replacement transaction

#### Scenario: Mapping response is invalid
- **WHEN** the Gemini response omits an input value, duplicates an input value, returns an unknown category, or returns an unexpected primary type display name
- **THEN** the stage MUST reject the response and avoid replacing the existing mapping table contents

### Requirement: Spot category mapping table
The system SHALL store generated category mappings in SQLite without changing raw geocode storage.

#### Scenario: Schema initialization creates mapping table
- **WHEN** the database schema is initialized
- **THEN** the schema MUST include a `spot_category_mappings` table keyed by `primary_type_display_name`
- **AND** each row MUST include `category_name`
- **AND** each row MAY include mapping diagnostics such as confidence, reason, model, and mapped timestamp
- **AND** `gemini_direct_spot_google_geocodes.primary_type_display_name` MUST remain unchanged as the raw Google Places value

### Requirement: Exported spots include category name
Each exported spot record SHALL include a tourist-friendly `categoryName` derived from the stored primary type display name mapping.

#### Scenario: Stored mapping exists
- **WHEN** an exportable Gemini geocode row has `primary_type_display_name` and `spot_category_mappings` contains a row for that value
- **THEN** the exported spot SHALL include `categoryName` set to the mapped category name
- **AND** the exported spot SHALL NOT include the raw primary type display name field

#### Scenario: Stored mapping is absent
- **WHEN** an exportable Gemini geocode row has no matching row in `spot_category_mappings`
- **THEN** the exported spot SHALL include `categoryName` set to `Overig`
- **AND** the spot SHALL remain exportable

#### Scenario: Primary type display name is absent
- **WHEN** an exportable Gemini geocode row has no `primary_type_display_name`
- **THEN** the exported spot SHALL include `categoryName` set to `Overig`
- **AND** the spot SHALL remain exportable

### Requirement: JSON export includes top-level filter metadata
The JSON export process SHALL include top-level `presenters` and `categories` arrays derived from exported, non-hidden spot records.

#### Scenario: Export contains presenter and category metadata
- **WHEN** export-data writes a JSON document with exportable non-hidden spots
- **THEN** the JSON SHALL contain a top-level `presenters` array of objects with `presenterName`
- **AND** the JSON SHALL contain a top-level `categories` array of objects with `categoryName`
- **AND** each top-level presenter and category value SHALL be represented by at least one exported spot

#### Scenario: Export has no spots
- **WHEN** export-data writes a JSON document with no exportable non-hidden spots
- **THEN** the JSON SHALL contain a valid top-level `spots` array
- **AND** the top-level `presenters` array SHALL be empty or omitted
- **AND** the top-level `categories` array SHALL be empty or omitted

## MODIFIED Requirements

### Requirement: Deterministic ordering in output
The system SHALL produce deterministic ordering of exported arrays to make output stable across runs with unchanged source data. Exported spots SHALL retain stable ordering by their existing deterministic spot fields. Exported presenters SHALL be ordered by the most recent associated article publication timestamp descending, with deterministic name tie-breaking. Exported categories SHALL be ordered by descending exported spot count, with deterministic name tie-breaking, except `Overig` SHALL always appear last when present.

#### Scenario: Repeated export with unchanged data
- **WHEN** the export is run multiple times against unchanged source data
- **THEN** the resulting JSON content MUST preserve consistent array ordering

#### Scenario: Presenter metadata is ordered by article recency
- **WHEN** exported spots contain multiple presenter names with associated article publication timestamps
- **THEN** the top-level `presenters` array MUST order presenters by their most recent associated article publication timestamp descending
- **AND** presenters without publication timestamps MUST appear after presenters with timestamps
- **AND** presenter name ordering MUST provide deterministic tie-breaking

#### Scenario: Category metadata is ordered by exported spot count
- **WHEN** exported spots contain multiple category names
- **THEN** the top-level `categories` array MUST order categories by descending exported spot count
- **AND** category name ordering MUST provide deterministic tie-breaking for equal counts
- **AND** `Overig` MUST appear last when present regardless of count

### Requirement: Valid JSON output for empty or partial datasets
The system SHALL always write valid JSON output, including cases where source data is empty or partially available.

#### Scenario: No matching rows in source data
- **WHEN** the source dataset has no exportable rows
- **THEN** the system MUST write a valid JSON document with an empty `spots` array
- **AND** any top-level `presenters` or `categories` arrays MUST be empty or omitted

### Requirement: JSON export supports Gemini spot source
The JSON export process SHALL read exportable geocoded spots, article links, presenter attribution, correction data, and stored category mappings from the Gemini-direct table set.

#### Scenario: Export invoked with default source
- **WHEN** a user runs `export-data`
- **THEN** the command SHALL read Gemini-derived geocode/link rows, Gemini-specific correction rows, and stored category mapping rows
- **AND** the generated JSON SHALL include a `spots` array with top-level `presenters` and `categories` metadata arrays
- **AND** each exported spot SHALL include presenter attribution derived from stored article presenter links
- **AND** each exported spot SHALL include `categoryName` derived from `spot_category_mappings` or `Overig` fallback behavior
- **AND** the generated JSON SHALL NOT include the raw primary type display name field on exported spots

### Requirement: JSON export omits top-level presenters
The JSON export process SHALL include a top-level `presenters` array because presenter filter ordering is part of the static JSON contract.

#### Scenario: Export contains presenters metadata
- **WHEN** export-data writes the frontend JSON document
- **THEN** the JSON SHALL contain a top-level `spots` array
- **AND** the JSON SHALL contain a top-level `presenters` array of objects with `presenterName`
- **AND** presenter names SHALL remain available on exported spot records for denormalized spot display and filtering

### Requirement: Exported spots include primary type display name
Each exported spot record SHALL include a mapped `categoryName` instead of the Google Places primary type display name.

#### Scenario: Primary type display name is stored
- **WHEN** an exportable Gemini geocode row has `primary_type_display_name`
- **THEN** the exported spot SHALL include a mapped `categoryName` in the spot record
- **AND** the exported spot SHALL NOT include the raw primary type display name field

#### Scenario: Primary type display name is absent
- **WHEN** an exportable Gemini geocode row has no `primary_type_display_name`
- **THEN** the exported spot SHALL remain exportable with `categoryName` set to `Overig`
- **AND** the exported spot SHALL NOT include the raw primary type display name field
