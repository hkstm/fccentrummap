## Context

The current export reads Google Places `primary_type_display_name` values from `gemini_direct_spot_google_geocodes` and publishes them directly in `spots.json`. These values are useful internal geocoding detail but are too specific for tourist-oriented display and filtering, for example splitting food places into many cuisine-specific restaurant types and shops into many narrow retail types.

The legacy `spots` table in `data/spots.db` is currently empty and does not contain `primary_type_display_name`; the active export source is the Gemini-direct geocode table. A database review found 190 Gemini geocode rows with non-empty primary type display names. The database schema should continue storing the raw Google value, while the public JSON export should publish a broader `categoryName` value.

## Goals / Non-Goals

**Goals:**

- Replace the public exported primary type display name field with `categoryName`.
- Use stored mappings from Google Places primary type display names to tourist-friendly categories during export.
- Keep category names stable and readable for frontend display/filtering.
- Restore top-level `presenters` filter metadata in `spots.json`.
- Add top-level `categories` filter metadata in `spots.json`.
- Keep raw `primary_type_display_name` storage unchanged in SQLite.
- Add a separate category-mapping stage that regenerates a stored SQLite mapping table from observed primary type display names.
- Make `export-data` use the stored mapping table rather than calling Gemini directly.
- Provide a clear fallback for missing or previously unseen primary type display names.

**Non-Goals:**

- Do not migrate or rewrite stored geocode rows.
- Do not change the Google Places geocoding stage or captured geocode payload.
- Do not make `export-data` perform network calls.
- Do not add a new external dependency beyond the existing Gemini API client.
- Do not create per-language/i18n category handling as part of this change.

## Decisions

### Separate category-mapping stage

Add a separate pipeline stage, for example `map-spot-categories`, that regenerates a stored SQLite mapping table from observed `primary_type_display_name` values. The stage should read distinct primary type display names from `gemini_direct_spot_google_geocodes`, call Gemini with a structured request, validate the response, and replace the table contents with the new complete mapping snapshot.

`export-data` should not call Gemini. It should read exportable spots and join or look up the stored category mapping table to produce each spot's `categoryName`.

Rationale: the desired categories are product/display taxonomy, not source data, but the mapping should be inspectable, stable between exports, and refreshable on demand. A separate stage keeps network/model behavior out of export, makes `spots.json` generation deterministic, and allows maintainers to rerun category mapping only when new primary type display names appear or taxonomy rules change.

Alternative considered: call Gemini during `export-data`. This was rejected because export should remain deterministic and should not depend on network/model availability. Another alternative was to maintain the mapping fully by hand in code; this was rejected because the observed primary type display names may change over time and the classification task is a good fit for a constrained structured model call.

### Mapping table

Add a SQLite table for generated category mappings. Exact naming can follow repository conventions; intended shape:

```sql
CREATE TABLE IF NOT EXISTS spot_category_mappings (
  primary_type_display_name TEXT PRIMARY KEY,
  category_name TEXT NOT NULL,
  confidence REAL,
  reason TEXT,
  model TEXT,
  mapped_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

The table stores one row per distinct primary type display name. The source geocode table remains unchanged. The mapping stage should replace the mapping table contents as a complete snapshot: run the Gemini mapping for the currently observed distinct primary type display names inside a transaction, delete existing rows, then insert the new validated result set. This avoids partial incremental updates and makes the table represent the latest complete mapping run.

### Structured Gemini category-mapping call

The category-mapping stage's API call should follow the existing structured Gemini pattern used by `geminidirect`: temperature `0`, `ResponseMIMEType: "application/json"`, a concrete `ResponseSchema`, and `GenerateContent`/`GenerateContentWithParts` through `internal/genai.Client`. No URL context or video parts are needed for category mapping; a text-only prompt is sufficient.

Example intended Go shape:

```go
package spotcategories

import (
	"context"

	gogenai "google.golang.org/genai"

	"fccentrummap/internal/genai"
)

type MappingInput struct {
	PrimaryTypeDisplayNames []string
}

func GenerateContentConfig() *gogenai.GenerateContentConfig {
	temperature := float32(0)
	return &gogenai.GenerateContentConfig{
		Temperature:      &temperature,
		ResponseMIMEType: "application/json",
		ResponseSchema:   responseSchema(),
		ThinkingConfig:   &gogenai.ThinkingConfig{IncludeThoughts: false},
	}
}

func responseSchema() *gogenai.Schema {
	return &gogenai.Schema{
		Type: gogenai.TypeObject,
		Properties: map[string]*gogenai.Schema{
			"mappings": {
				Type: gogenai.TypeArray,
				Items: &gogenai.Schema{
					Type: gogenai.TypeObject,
					Properties: map[string]*gogenai.Schema{
						"primaryTypeDisplayName": {Type: gogenai.TypeString},
						"categoryName":           {Type: gogenai.TypeString},
						"confidence":             {Type: gogenai.TypeNumber},
						"reason":                 {Type: gogenai.TypeString},
					},
					Required: []string{"primaryTypeDisplayName", "categoryName", "confidence", "reason"},
				},
			},
		},
		Required: []string{"mappings"},
	}
}

func GenerateMapping(ctx context.Context, client *genai.Client, input MappingInput) ([]byte, error) {
	prompt := BuildPrompt(input)
	result, err := client.GenerateContent(ctx, prompt, GenerateContentConfig())
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}
```

Before calling Gemini, the mapping stage should trim and deduplicate the observed `primary_type_display_name` values so each distinct primary type display name is classified only once per mapping generation.

The mapping stage should validate the model response before replacing the table contents:

- every requested distinct `primaryTypeDisplayName` appears exactly once;
- every returned `categoryName` is one of the fixed categories;
- duplicate or unknown primary type display names are rejected;
- invalid or missing mappings fall back to `Overig` only after validation identifies them as unmappable, not because parsing failed silently.

After validation, the stage should replace the contents of `spot_category_mappings` with the new complete mapping result. `export-data` should use `COALESCE(spot_category_mappings.category_name, 'Overig')` or equivalent application logic so newly observed but unmapped primary type display names export as `Overig` until the mapping stage is rerun.

Full English prompt template:

```text
You are an assistant that maps Google Places primary type display names to fixed categories for a tourist map of Amsterdam.

Important:
- The category labels are fixed and are mostly Dutch.
- The primary type display names are Google Places labels and are mostly Dutch, but they may sometimes be English.
- For each primary type display name, choose exactly one category from the fixed list.
- Use only categories from the fixed list; do not invent new categories.
- Choose the category based on what a tourist is likely looking for, such as food, shopping, museums/culture, sightseeing, accommodation, nature, nightlife, or practical services.
- Use `Overig` only when none of the fixed categories is a reasonable fit.
- The input list has already been deduplicated; return exactly one mapping for every input value.
- Preserve each input value exactly in `primaryTypeDisplayName`.

Fixed categories:
- Restaurants
- Cafés & Bakkerijen
- Bars & Nachtleven
- Winkelen
- Boodschappen & Markten
- Musea & Cultuur
- Bezienswaardigheden
- Verblijf
- Beauty & Wellness
- Natuur
- Praktisch & Diensten
- Overig

Primary type display names to classify:
{{PRIMARY_TYPE_DISPLAY_NAMES_JSON_ARRAY}}

Return only JSON matching exactly this shape:
{
  "mappings": [
    {
      "primaryTypeDisplayName": "<exact input value>",
      "categoryName": "<one fixed category>",
      "confidence": 0.0,
      "reason": "<short English explanation>"
    }
  ]
}
```


### Tourist-friendly category set

Use broad Dutch categories that match common visitor intent:

- `Restaurants`
- `Cafés & Bakkerijen`
- `Bars & Nachtleven`
- `Winkelen`
- `Boodschappen & Markten`
- `Musea & Cultuur`
- `Bezienswaardigheden`
- `Verblijf`
- `Beauty & Wellness`
- `Natuur`
- `Praktisch & Diensten`
- `Overig`

The initial mapping should cover the primary type display names observed in `data/spots.db`, including restaurants by cuisine, cafes/bakeries, retail shops, grocery/market locations, cultural venues, landmarks, hotels, wellness services, parks, and administrative/community/service locations.

Alternative considered: expose the raw Google type and let the frontend group it. This was rejected because it would duplicate taxonomy logic in the frontend and keep the public static data contract tied to over-specific source labels.

### Fallback behavior

If `primary_type_display_name` is missing or no row exists in `spot_category_mappings`, export `categoryName` as `Overig`.

Rationale: every exported spot can still participate in category display/filtering, and newly observed unmapped Google labels are visible as a bounded fallback rather than breaking export or producing inconsistent null handling. The presence of `Overig` also makes it clear that the mapping stage may need to be rerun or reviewed.

Alternative considered: omit `categoryName` or set it to null for unknown values. This was rejected because it complicates frontend filter derivation and creates a less useful public contract.

### Top-level filter metadata

Restore a top-level `presenters` array and add a top-level `categories` array to `spots.json`. Spot records should remain denormalized with their own `presenterName` and `categoryName`, while the top-level arrays provide explicit ordered filter options for the frontend.

Presenter ordering should be based on recency: sort distinct exported presenters by the most recent `article_sources.published_at` value for an article associated with that presenter, descending. If a presenter has no published timestamp, place them after presenters with timestamps and use `presenterName` as a deterministic tie-breaker.

Category ordering should be based on popularity in the exported spot set: sort distinct exported categories by descending spot count, then by `categoryName` as a deterministic tie-breaker. Always place `Overig` last when it is present, regardless of its count.

### Intended `spots.json` schema

The exported JSON document should have this public shape:

```json
{
  "presenters": [
    {
      "presenterName": "Foo"
    }
  ],
  "categories": [
    {
      "categoryName": "Beauty & Wellness"
    }
  ],
  "spots": [
    {
      "spotId": "123:456",
      "placeId": "places/example",
      "spotName": "Example Spot",
      "presenterName": "Foo",
      "categoryName": "Beauty & Wellness",
      "latitude": 52.3676,
      "longitude": 4.9041,
      "youtubeLink": "https://www.youtube.com/watch?v=example&t=42s",
      "articleUrl": "https://fccentrum.nl/example"
    }
  ]
}
```

`presenters` and `categories` should contain only values represented by exported, non-hidden spots. They are metadata for filter ordering and should not replace the denormalized fields on each spot.

### Public contract rename

The exported spot struct/JSON should remove the old primary type display name field and add `categoryName`.

Rationale: the field name should describe the public contract, not the implementation source. This is intentionally breaking for consumers of `spots.json` and should be reflected in specs and frontend updates.

Alternative considered: keep both fields temporarily. This was rejected to avoid preserving the over-specific display field in the public contract and because the generated JSON is an internal static-site build artifact.

## Risks / Trade-offs

- Unmapped future Google types may be grouped as `Overig` → Add focused unit tests for known mappings and make the mapping table easy to regenerate.
- Category choices are subjective → Start with broad tourist intent categories and adjust based on actual UI/filter needs.
- Breaking JSON field rename may break frontend code → Update frontend types/usages and regenerate or adjust sample/static data in the same implementation.
- Top-level `presenters` and `categories` arrays can drift from spot records if assembled separately → Derive both arrays from the final exported, corrected, non-hidden spot set.
- Category labels are Dutch while some Google source labels are English → Treat exported category labels as the frontend display taxonomy; defer broader localization until needed.
