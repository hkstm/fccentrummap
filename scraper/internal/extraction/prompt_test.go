package extraction

import (
	"strings"
	"testing"
)

func TestBuildDutchPass1PromptContainsRequiredGuidance(t *testing.T) {
	prompt, err := BuildDutchPass1Prompt(PromptInput{
		CleanedArticleText: "Niels laat de Stopera en De Pijp zien als favoriete spots.",
		Sentences:          []SentenceUnit{{Text: "We lopen door De Pijp", Start: 12.5}},
	})
	if err != nil {
		t.Fatalf("BuildDutchPass1Prompt() error = %v", err)
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

func TestBuildDutchPass2RefinementPromptContainsPass1AndSchema(t *testing.T) {
	orig1 := 15.0
	orig2 := 33.2
	prompt, err := BuildDutchPass2RefinementPrompt(RefinementPromptInput{
		Sentences: []SentenceUnit{{Text: "We starten bij Stopera", Start: 10.2}},
		Pass1Spots: []Candidate{
			{Place: "Stopera", OriginalSentenceStartTimestamp: &orig1},
			{Place: "Oosterpark", OriginalSentenceStartTimestamp: &orig2},
		},
	})
	if err != nil {
		t.Fatalf("BuildDutchPass2RefinementPrompt() error = %v", err)
	}

	for _, snippet := range []string{
		"[pass1_spots]",
		"[transcript_sentences]",
		"refinedSentenceStartTimestamp",
		"<=",
		"submit_refined_spots",
		"place=Stopera",
	} {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("pass-2 prompt is missing expected snippet %q", snippet)
		}
	}
}
