## 1. Frontend workspace bootstrap

- [ ] 1.1 Create the frontend app structure under `viz/` with the required Next.js/TypeScript project files
- [ ] 1.2 Add baseline app shell, styling, and environment-variable handling for Google Maps configuration

## 2. Static data export and loading

- [ ] 2.1 Implement `scraper/cmd/export` integration for the frontend data contract at `viz/public/data/spots.json`
- [ ] 2.2 Implement frontend data loading and error handling for `/data/spots.json`

## 3. Map rendering

- [ ] 3.1 Render a Google Map centered on Amsterdam using the Advanced Marker API
- [ ] 3.2 Render one marker per exported spot using the static JSON dataset

## 4. Marker styling and filtering

- [ ] 4.1 Implement Amsterdam andreaskruis marker rendering with a deterministic color palette
- [ ] 4.2 Implement the author filter panel with default selection, toggles, and collapse behavior

## 5. Verification

- [ ] 5.1 Verify the build-time export -> static JSON -> frontend render flow end-to-end
- [ ] 5.2 Verify missing-config and missing-data failure states are surfaced clearly
