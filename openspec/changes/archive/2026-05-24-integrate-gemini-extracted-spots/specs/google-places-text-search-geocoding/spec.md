## REMOVED Requirements

### Requirement: Geocode stage is file-mode only in this change
**Reason**: SQLite integration is now required so Gemini-derived and transcript-derived spot mentions can be geocoded from the selected database source.

**Migration**: Use `geocode-spots --spot-source gemini-direct` or omit `--spot-source` for the default Gemini source, and use `geocode-spots --spot-source transcript` to retain transcript-derived behavior. File mode remains available for explicit artifact processing where already supported.

## ADDED Requirements

### Requirement: Geocode stage supports source-selected SQLite geocoding
The `geocode-spots` stage SHALL geocode spot mentions from the selected SQLite spot source and SHALL write geocode and article-link rows to that same source's table set.

#### Scenario: Geocode invoked in SQLite mode with default source
- **WHEN** a user runs `geocode-spots` in SQLite mode without `--spot-source`
- **THEN** the command SHALL select `gemini-direct`
- **AND** it SHALL read ungeocoded rows from `gemini_direct_spot_mentions`
- **AND** it SHALL write successful geocodes to `gemini_direct_spot_google_geocodes`
- **AND** it SHALL write article links to `gemini_direct_article_spots`

#### Scenario: Geocode invoked in SQLite mode with transcript source
- **WHEN** a user runs `geocode-spots` in SQLite mode with `--spot-source transcript`
- **THEN** the command SHALL read ungeocoded rows from transcript-derived `spot_mentions`
- **AND** it SHALL write successful geocodes to `spot_google_geocodes`
- **AND** it SHALL write article links to `article_spots`
- **AND** it SHALL NOT write Gemini-specific geocode or article-link rows

#### Scenario: Geocode invoked with unsupported spot source
- **WHEN** a user runs `geocode-spots` with an unsupported `--spot-source` value
- **THEN** validation SHALL fail before Google Places requests are issued
- **AND** the error SHALL name `gemini-direct` and `transcript` as supported values

### Requirement: Geocode inline export uses the selected spot source
When `geocode-spots` performs an inline JSON export, the export step SHALL use the same selected spot source as the geocode step.

#### Scenario: Gemini geocode with inline export
- **WHEN** `geocode-spots` runs with `--spot-source gemini-direct` and inline export enabled
- **THEN** the geocode step SHALL write Gemini-specific geocode/link rows
- **AND** the inline export SHALL read Gemini-specific geocode/link/correction rows
- **AND** the inline export SHALL NOT read transcript-derived spot rows for that run

#### Scenario: Transcript geocode with inline export
- **WHEN** `geocode-spots` runs with `--spot-source transcript` and inline export enabled
- **THEN** the geocode step SHALL write transcript-derived geocode/link rows
- **AND** the inline export SHALL read transcript-derived geocode/link/correction rows
- **AND** the inline export SHALL NOT read Gemini-specific spot rows for that run

### Requirement: Geocode stage preserves file-mode artifact processing
The `geocode-spots` stage SHALL continue to support explicit file-mode geocoding for artifact-based processing where that mode is selected.

#### Scenario: Geocode invoked in file mode
- **WHEN** a user runs `geocode-spots --io file --in <path>`
- **THEN** the stage SHALL process the explicit input artifact
- **AND** it SHALL emit deterministic geocode output artifact(s)
- **AND** source-selected SQLite table writes SHALL NOT be required for the file-mode run
