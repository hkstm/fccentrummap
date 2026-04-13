## ADDED Requirements

### Requirement: Geocode spot addresses to coordinates
The geocoder SHALL resolve each spot address to latitude and longitude using the Google Maps Geocoding API. It SHALL append ", Amsterdam" to each address to improve accuracy.

#### Scenario: Successful geocoding
- **WHEN** a spot with address "Amstel 3" is geocoded
- **THEN** the geocoder SHALL return valid latitude and longitude coordinates

#### Scenario: Address not found
- **WHEN** the Geocoding API returns no results for an address
- **THEN** the geocoder SHALL treat this as a geocoding failure

### Requirement: Fail-fast on geocoding errors
The geocoder SHALL abort the entire run immediately if any geocoding request fails. It SHALL NOT attempt to geocode any remaining spots.

#### Scenario: API error during geocoding
- **WHEN** a geocoding request fails (network error, invalid API key, quota exhausted, or no results)
- **THEN** the program SHALL log an error including the spot name, address, and the API error message, and exit without processing any further spots

#### Scenario: Already-completed articles are preserved
- **WHEN** the geocoder aborts due to an error
- **THEN** articles already marked `COMPLETED` in `articles_raw` SHALL remain unchanged

#### Scenario: Current article stays PENDING
- **WHEN** the geocoder aborts while processing an article
- **THEN** that article SHALL remain `PENDING` in `articles_raw` so it can be retried on the next run

### Requirement: Google Maps API key configuration
The geocoder SHALL read the Google Maps API key from an environment variable.

#### Scenario: API key provided
- **WHEN** the environment variable `GOOGLE_MAPS_API_KEY` is set
- **THEN** the geocoder SHALL use it for all Geocoding API requests

#### Scenario: API key missing
- **WHEN** the environment variable `GOOGLE_MAPS_API_KEY` is not set
- **THEN** the program SHALL log an error stating the required environment variable is missing, and exit before attempting any geocoding
