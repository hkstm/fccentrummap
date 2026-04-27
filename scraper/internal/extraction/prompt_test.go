package extraction

import (
	"strings"
	"testing"
)

func TestBuildDutchPromptContainsRequiredGuidance(t *testing.T) {
	prompt, err := BuildDutchPrompt(PromptInput{
		CleanedArticleText: "Niels laat de Stopera en De Pijp zien als favoriete spots.",
		Sentences:          []SentenceUnit{{Text: "We lopen door De Pijp", Start: 12.5}},
	})
	if err != nil {
		t.Fatalf("BuildDutchPrompt() error = %v", err)
	}

	expectedSnippets := []string{
		"De spots van",
		"[cleaned_article]",
		"[transcript_sentences]",
		"primaire bron",
		"transcriptbewijs",
		"presenter_name",
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
