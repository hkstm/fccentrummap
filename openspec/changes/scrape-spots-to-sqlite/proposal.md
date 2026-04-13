## Why

FC Centrum publishes articles where Amsterdam locals recommend their favorite spots in the city center. This data is scattered across paginated web pages with no structured access. We need to scrape, geocode, and store this data in SQLite so it can later power a Google Maps-based interactive map of recommended spots.

## What Changes

- Build a Go scraper that crawls all paginated "spots" articles from `fccentrum.nl/categorie/spots/`
- Parse each article to extract the author name and their listed spots (name + address)
- Geocode each spot address to lat/lng coordinates using the Google Maps Geocoding API
- Store everything in a SQLite database with normalized tables for authors, spots, and their relationships

## Capabilities

### New Capabilities
- `web-scraper`: Crawl the fccentrum.nl spots category pages and individual article pages, extract article metadata (author, date, URL) and spot listings (name, address) from the HTML content
- `geocoding`: Resolve spot addresses to geographic coordinates (latitude/longitude) using the Google Maps Geocoding API
- `sqlite-storage`: SQLite database schema and data access layer with tables for authors, spots, and author-spot relationships

### Modified Capabilities

## Impact

- New Go module with dependencies: `colly` (scraping framework), `modernc.org/sqlite` (pure Go, no CGo), `googlemaps`
- Requires a Google Maps API key with Geocoding API enabled
- Creates a `data/spots.db` SQLite database file for the scraper pipeline
- No existing code is affected (greenfield project)
