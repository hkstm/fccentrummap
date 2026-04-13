## MODIFIED Requirements

### Requirement: Initialize schema on startup
The scraper repository layer SHALL create required tables if they do not already exist.

#### Scenario: Fresh database
- **WHEN** the program opens a new `data/spots.db`
- **THEN** it SHALL create `articles_raw`, `authors`, `spots`, `articles`, and `article_spots`

#### Scenario: Existing database
- **WHEN** the program opens an existing database with the required schema
- **THEN** existing data SHALL be preserved

### Requirement: Export query supports frontend JSON generation
The repository SHALL provide a query that joins spots, articles, and authors into export-ready data.

#### Scenario: Exporting current map data
- **WHEN** the exporter requests data from the repository
- **THEN** the repository SHALL return deduplicated spot records with their coordinates and associated author names

### Requirement: Spot storage supports transcript-derived video metadata
The repository SHALL support storing optional source video metadata for transcript-derived spots.

#### Scenario: Spot with transcript source
- **WHEN** a spot is stored from a transcript-first extraction flow
- **THEN** the persistence layer SHALL support storing optional `video_url` and `timestamp_seconds` values for that spot
