package spotcategories

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/hkstm/fccentrummap/internal/genai"
	"github.com/hkstm/fccentrummap/internal/repository"
	gogenai "google.golang.org/genai"
)

const FallbackCategory = "Overig"

var Categories = []string{
	"Restaurants",
	"Cafés & Bakkerijen",
	"Bars & Nachtleven",
	"Winkelen",
	"Boodschappen & Markten",
	"Musea & Cultuur",
	"Bezienswaardigheden",
	"Verblijf",
	"Beauty & Wellness",
	"Natuur",
	"Praktisch & Diensten",
	FallbackCategory,
}

type MappingInput struct {
	PrimaryTypeDisplayNames []string
}

type Mapping struct {
	PrimaryTypeDisplayName string  `json:"primaryTypeDisplayName"`
	CategoryName           string  `json:"categoryName"`
	Confidence             float64 `json:"confidence"`
	Reason                 string  `json:"reason"`
}

type response struct {
	Mappings []Mapping `json:"mappings"`
}

type generateContentEnvelope struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text    string `json:"text"`
				Thought bool   `json:"thought"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type Repository interface {
	ListDistinctPrimaryTypeDisplayNames() ([]string, error)
	ReplaceSpotCategoryMappings([]repository.SpotCategoryMapping) error
}

func NormalizePrimaryTypeDisplayNames(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	slices.Sort(out)
	return out
}

func IsKnownCategory(category string) bool {
	return slices.Contains(Categories, strings.TrimSpace(category))
}

func BuildPrompt(input MappingInput) string {
	names := NormalizePrimaryTypeDisplayNames(input.PrimaryTypeDisplayNames)
	payload, _ := json.Marshal(names)
	var b strings.Builder
	b.WriteString("You are an assistant that maps Google Places primary type display names to fixed categories for a tourist map of Amsterdam.\n\n")
	b.WriteString("Important:\n")
	b.WriteString("- The category labels are fixed and are mostly Dutch.\n")
	b.WriteString("- The primary type display names are Google Places labels and are mostly Dutch, but they may sometimes be English.\n")
	b.WriteString("- For each primary type display name, choose exactly one category from the fixed list.\n")
	b.WriteString("- Use only categories from the fixed list; do not invent new categories.\n")
	b.WriteString("- Choose the category based on what a tourist is likely looking for, such as food, shopping, museums/culture, sightseeing, accommodation, nature, nightlife, or practical services.\n")
	b.WriteString("- Use `Overig` only when none of the fixed categories is a reasonable fit.\n")
	b.WriteString("- The input list has already been deduplicated; return exactly one mapping for every input value.\n")
	b.WriteString("- Preserve each input value exactly in `primaryTypeDisplayName`.\n\n")
	b.WriteString("Fixed categories:\n")
	for _, category := range Categories {
		b.WriteString("- ")
		b.WriteString(category)
		b.WriteString("\n")
	}
	b.WriteString("\nPrimary type display names to classify:\n")
	b.Write(payload)
	b.WriteString("\n\nReturn only JSON matching exactly this shape:\n")
	b.WriteString("{\n  \"mappings\": [\n    {\n      \"primaryTypeDisplayName\": \"<exact input value>\",\n      \"categoryName\": \"<one fixed category>\",\n      \"confidence\": 0.0,\n      \"reason\": \"<short English explanation>\"\n    }\n  ]\n}\n")
	return b.String()
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
	result, err := client.GenerateContent(ctx, BuildPrompt(input), GenerateContentConfig())
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func extractStructuredResponse(rawBody []byte, target *response) error {
	if err := unmarshalDirectStructuredJSON(rawBody, target); err == nil {
		return nil
	}

	var envelope generateContentEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("model response is not valid generateContent JSON: %w", err)
	}
	var lastErr error
	for _, cand := range envelope.Candidates {
		for _, part := range cand.Content.Parts {
			if part.Thought || strings.TrimSpace(part.Text) == "" {
				continue
			}
			if err := json.Unmarshal([]byte(strings.TrimSpace(part.Text)), target); err != nil {
				lastErr = err
				continue
			}
			return nil
		}
	}
	if lastErr != nil {
		return fmt.Errorf("model structured response text is not valid category mapping JSON: %w", lastErr)
	}
	return fmt.Errorf("model response does not contain category mapping structured JSON output")
}

func unmarshalDirectStructuredJSON(rawBody []byte, target *response) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &fields); err != nil {
		return err
	}
	if _, ok := fields["mappings"]; !ok {
		return fmt.Errorf("top-level mappings field not found")
	}
	return json.Unmarshal(rawBody, target)
}

func ParseAndValidate(body []byte, requested []string) ([]Mapping, error) {
	var r response
	if err := extractStructuredResponse(body, &r); err != nil {
		return nil, err
	}
	expected := NormalizePrimaryTypeDisplayNames(requested)
	expectedSet := map[string]bool{}
	for _, name := range expected {
		expectedSet[name] = true
	}
	seen := map[string]bool{}
	out := make([]Mapping, 0, len(r.Mappings))
	for _, mapping := range r.Mappings {
		primary := strings.TrimSpace(mapping.PrimaryTypeDisplayName)
		category := strings.TrimSpace(mapping.CategoryName)
		if primary == "" {
			return nil, fmt.Errorf("mapping response contains empty primaryTypeDisplayName")
		}
		if !expectedSet[primary] {
			return nil, fmt.Errorf("mapping response contains unexpected primaryTypeDisplayName %q", primary)
		}
		if seen[primary] {
			return nil, fmt.Errorf("mapping response duplicates primaryTypeDisplayName %q", primary)
		}
		if !IsKnownCategory(category) {
			return nil, fmt.Errorf("mapping response category %q for %q is not fixed", category, primary)
		}
		seen[primary] = true
		mapping.PrimaryTypeDisplayName = primary
		mapping.CategoryName = category
		mapping.Reason = strings.TrimSpace(mapping.Reason)
		out = append(out, mapping)
	}
	for _, primary := range expected {
		if !seen[primary] {
			return nil, fmt.Errorf("mapping response missing primaryTypeDisplayName %q", primary)
		}
	}
	return out, nil
}

type Service struct {
	repo   Repository
	client *genai.Client
	model  string
}

func NewService(repo Repository, client *genai.Client, model string) *Service {
	return &Service{repo: repo, client: client, model: strings.TrimSpace(model)}
}

func (s *Service) Run(ctx context.Context) (int, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("repository is required")
	}
	if s.client == nil {
		return 0, fmt.Errorf("Gemini client is required")
	}
	names, err := s.repo.ListDistinctPrimaryTypeDisplayNames()
	if err != nil {
		return 0, err
	}
	names = NormalizePrimaryTypeDisplayNames(names)
	if len(names) == 0 {
		return 0, s.repo.ReplaceSpotCategoryMappings(nil)
	}
	var mappings []Mapping
	const batchSize = 50
	for start := 0; start < len(names); start += batchSize {
		end := min(start+batchSize, len(names))
		batch := names[start:end]
		body, err := GenerateMapping(ctx, s.client, MappingInput{PrimaryTypeDisplayNames: batch})
		if err != nil {
			return 0, err
		}
		batchMappings, err := ParseAndValidate(body, batch)
		if err != nil {
			if len(batch) == 1 {
				return 0, err
			}
			for _, name := range batch {
				body, err := GenerateMapping(ctx, s.client, MappingInput{PrimaryTypeDisplayNames: []string{name}})
				if err != nil {
					return 0, err
				}
				singleMappings, err := ParseAndValidate(body, []string{name})
				if err != nil {
					return 0, err
				}
				mappings = append(mappings, singleMappings...)
			}
			continue
		}
		mappings = append(mappings, batchMappings...)
	}
	repoMappings := make([]repository.SpotCategoryMapping, 0, len(mappings))
	for _, mapping := range mappings {
		confidence := mapping.Confidence
		reason := mapping.Reason
		repoMappings = append(repoMappings, repository.SpotCategoryMapping{
			PrimaryTypeDisplayName: mapping.PrimaryTypeDisplayName,
			CategoryName:           mapping.CategoryName,
			Confidence:             &confidence,
			Reason:                 &reason,
			Model:                  s.model,
		})
	}
	if err := s.repo.ReplaceSpotCategoryMappings(repoMappings); err != nil {
		return 0, err
	}
	return len(repoMappings), nil
}
