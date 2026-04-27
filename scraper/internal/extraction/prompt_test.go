package extraction

import (
	"strings"
	"testing"
)

func TestBuildDutchPromptContainsRequiredGuidance(t *testing.T) {
	prompt, err := BuildDutchPrompt([]SentenceUnit{{Text: "We lopen door De Pijp", Start: 12.5}})
	if err != nil {
		t.Fatalf("BuildDutchPrompt() error = %v", err)
	}

	expectedSnippets := []string{
		"De spots van",
		"Amsterdam",
		"2 tot 7",
		"canonieke spot",
		"minimaal 2 verschillende zinnen",
		"zonder voorvoegsels",
		"place",
		"sentenceStartTimestamp",
		"submit_spots",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("prompt is missing expected snippet %q", snippet)
		}
	}
}
