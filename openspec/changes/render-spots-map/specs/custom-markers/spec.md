## ADDED Requirements

### Requirement: Markers display Amsterdam andreaskruis
Each spot marker SHALL render as an Amsterdam "X" (andreaskruis / St. Andrew's cross) shape using an inline SVG element as the `AdvancedMarker` content.

#### Scenario: Marker renders as andreaskruis
- **WHEN** a spot marker is displayed on the map
- **THEN** it SHALL show an andreaskruis (saltire/X shape) SVG instead of the default Google Maps pin

### Requirement: Marker color is determined by author
Each marker's fill color SHALL be assigned deterministically based on the author. All spots from the same author SHALL have the same color.

#### Scenario: Same author same color
- **WHEN** two spots belong to the same author
- **THEN** both markers SHALL have the same fill color

#### Scenario: Different authors get different colors
- **WHEN** two spots belong to different authors
- **THEN** their markers SHALL have different fill colors, unless the palette has wrapped (more authors than palette colors)

### Requirement: Color palette is fixed and visually distinct
The application SHALL define a fixed palette of 12-16 curated colors that are visually distinguishable from each other and from the map background.

#### Scenario: Color assignment from palette
- **WHEN** authors are assigned colors
- **THEN** each author SHALL receive a color from the palette based on a deterministic index (sorted author list, index mod palette length)

#### Scenario: Palette wraps for many authors
- **WHEN** there are more authors than palette colors
- **THEN** the palette SHALL wrap (cycle), and the assignment SHALL remain deterministic across page loads

### Requirement: Markers are legible at map zoom levels
The andreaskruis SVG SHALL be sized to remain visible and recognizable at typical city-level zoom levels without obscuring the map.

#### Scenario: Marker size at default zoom
- **WHEN** the map is at the default Amsterdam zoom level
- **THEN** markers SHALL be large enough to tap/click but small enough not to overlap excessively
