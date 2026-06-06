package spotcategories

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildPromptIncludesFixedCategoriesAndInput(t *testing.T) {
	prompt := BuildPrompt(MappingInput{PrimaryTypeDisplayNames: []string{" Café ", "Restaurant", "Café"}})
	if !strings.Contains(prompt, "Cafés & Bakkerijen") || !strings.Contains(prompt, "Overig") {
		t.Fatalf("prompt missing fixed categories: %s", prompt)
	}
	if !strings.Contains(prompt, `["Café","Restaurant"]`) {
		t.Fatalf("prompt should include sorted deduplicated JSON input, got: %s", prompt)
	}
}

func TestParseAndValidateRejectsInvalidResponses(t *testing.T) {
	valid := response{Mappings: []Mapping{{PrimaryTypeDisplayName: "Café", CategoryName: "Cafés & Bakkerijen", Confidence: 0.9, Reason: "coffee"}}}
	body, _ := json.Marshal(valid)
	if _, err := ParseAndValidate(body, []string{"Café"}); err != nil {
		t.Fatalf("valid response rejected: %v", err)
	}

	cases := map[string]response{
		"missing":     {Mappings: nil},
		"duplicate":   {Mappings: []Mapping{{PrimaryTypeDisplayName: "Café", CategoryName: "Cafés & Bakkerijen"}, {PrimaryTypeDisplayName: "Café", CategoryName: "Cafés & Bakkerijen"}}},
		"unknown cat": {Mappings: []Mapping{{PrimaryTypeDisplayName: "Café", CategoryName: "Bad"}}},
		"unexpected":  {Mappings: []Mapping{{PrimaryTypeDisplayName: "Hotel", CategoryName: "Verblijf"}}},
	}
	for name, tc := range cases {
		body, _ := json.Marshal(tc)
		if _, err := ParseAndValidate(body, []string{"Café"}); err == nil {
			t.Fatalf("%s response accepted", name)
		}
	}
}
