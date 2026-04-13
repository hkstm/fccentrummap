## ADDED Requirements

### Requirement: Initialize database schema on startup
The repository SHALL create all tables if they do not already exist when the program starts.

#### Scenario: Fresh database
- **WHEN** the program runs and `data/spots.db` does not exist
- **THEN** the database SHALL be created with all tables: `articles_raw`, `authors`, `spots`, `articles`, `article_spots`

#### Scenario: Existing database
- **WHEN** the program runs and `data/spots.db` already exists with the correct schema
- **THEN** the existing data SHALL be preserved

### Requirement: Store raw article HTML
The repository SHALL insert fetched HTML into the `articles_raw` table with status `PENDING`.

#### Scenario: Insert new raw article
- **WHEN** a new article URL and HTML are provided
- **THEN** a row SHALL be inserted into `articles_raw` with the URL, HTML, status `PENDING`, and current timestamps

#### Scenario: Duplicate URL
- **WHEN** an article URL already exists in `articles_raw`
- **THEN** the insert SHALL be skipped (no error)

### Requirement: Store parsed article data
The repository SHALL insert parsed data into `authors`, `articles`, `spots`, and `article_spots` tables when processing succeeds.

#### Scenario: New author
- **WHEN** an author name does not exist in `authors`
- **THEN** a new row SHALL be inserted

#### Scenario: Existing author
- **WHEN** an author name already exists in `authors`
- **THEN** the existing `author_id` SHALL be reused

#### Scenario: New spot with coordinates
- **WHEN** a spot with name, address, latitude, and longitude is stored
- **THEN** a row SHALL be inserted into `spots` with all fields populated

#### Scenario: Duplicate spot
- **WHEN** a spot with the same name and address already exists
- **THEN** the existing `spot_id` SHALL be reused

#### Scenario: Link article to spots
- **WHEN** an article is successfully processed
- **THEN** rows SHALL be inserted into `article_spots` linking the `article_id` to each `spot_id`

### Requirement: Update articles_raw status
The repository SHALL update the status of `articles_raw` entries as processing progresses.

#### Scenario: Processing succeeds
- **WHEN** an article is fully parsed and all its spots are geocoded
- **THEN** the `articles_raw` status SHALL be set to `COMPLETED` and `updated_at` SHALL be refreshed

#### Scenario: Processing fails
- **WHEN** parsing or geocoding fails for an article
- **THEN** the `articles_raw` status SHALL be set to `FAILED` and `updated_at` SHALL be refreshed
- **AND** the repository SHALL log an error including the `article_raw_id`, URL, and the reason for failure

### Requirement: Query PENDING articles for processing
The repository SHALL provide a way to retrieve all `articles_raw` entries with status `PENDING`.

#### Scenario: PENDING articles exist
- **WHEN** there are entries with status `PENDING` in `articles_raw`
- **THEN** the repository SHALL return their `article_raw_id`, URL, and HTML

#### Scenario: No PENDING articles
- **WHEN** all entries are `COMPLETED` or `FAILED`
- **THEN** the repository SHALL return an empty result
