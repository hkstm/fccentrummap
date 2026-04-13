## ADDED Requirements

### Requirement: Map renders centered on Amsterdam
The application SHALL display a full-viewport Google Maps instance centered on Amsterdam (approximately 52.37°N, 4.90°E) at a zoom level that shows the city center.

#### Scenario: Initial map load
- **WHEN** the page loads
- **THEN** the map SHALL be centered on Amsterdam and display all spot markers within the city bounds

### Requirement: Map uses Advanced Marker API
The application SHALL use the Google Maps Advanced Marker API via `@vis.gl/react-google-maps` with a cloud-based Map ID.

#### Scenario: Map initializes with Map ID
- **WHEN** the map component mounts
- **THEN** it SHALL initialize with the Map ID from `NEXT_PUBLIC_GOOGLE_MAPS_MAP_ID` and the API key from `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY`

#### Scenario: Missing environment variables
- **WHEN** the Map ID or API key environment variable is not set
- **THEN** the application SHALL display an error message instead of the map

### Requirement: All spots render as markers
The application SHALL render one `AdvancedMarker` for each spot in the loaded data, positioned at the spot's latitude and longitude.

#### Scenario: Spots loaded successfully
- **WHEN** the static JSON data is loaded
- **THEN** one marker SHALL appear on the map for each spot at its geocoded position

#### Scenario: Spot with multiple authors
- **WHEN** a spot appears in articles by multiple authors
- **THEN** the spot SHALL render as a single marker (spots are unique by name+address)

### Requirement: Map is interactive
The user SHALL be able to pan and zoom the map freely using standard Google Maps controls.

#### Scenario: User pans the map
- **WHEN** the user drags the map
- **THEN** the map SHALL pan to follow the drag gesture

#### Scenario: User zooms the map
- **WHEN** the user scrolls or uses zoom controls
- **THEN** the map SHALL zoom in or out accordingly
