package extraction

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeterministicFixturesPromptAndResponse(t *testing.T) {
	article, err := os.ReadFile(filepath.Join("testdata", "prompt_input_article.txt"))
	if err != nil {
		t.Fatalf("read article fixture: %v", err)
	}

	prompt, err := BuildDutchPrompt(PromptInput{
		CleanedArticleText: string(article),
		Sentences: []SentenceUnit{
			{Text: "Vandaag laat ik de Stopera zien", Start: 14.5},
			{Text: "Daarna gaan we naar Oosterpark", Start: 33.2},
		},
	})
	if err != nil {
		t.Fatalf("build prompt from fixture: %v", err)
	}
	for _, required := range []string{"[cleaned_article]", "[transcript_sentences]", "presenter_name", "sentenceStartTimestamp"} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("prompt missing required token %q", required)
		}
	}

	withPresenter, err := os.ReadFile(filepath.Join("testdata", "response_with_presenter.json"))
	if err != nil {
		t.Fatalf("read response_with_presenter fixture: %v", err)
	}
	parsed, err := ParseAndValidateResponse(withPresenter)
	if err != nil {
		t.Fatalf("parse response_with_presenter fixture: %v", err)
	}
	if parsed.PresenterName == nil || *parsed.PresenterName != "Niels Oosthoek" {
		t.Fatalf("unexpected presenter_name from fixture: %+v", parsed.PresenterName)
	}
	if len(parsed.Spots) != 2 {
		t.Fatalf("expected 2 spots from fixture, got %d", len(parsed.Spots))
	}
	if parsed.Spots[0].OriginalSentenceStartTimestamp == nil || parsed.Spots[0].RefinedSentenceStartTimestamp == nil {
		t.Fatalf("expected fixture spot to include both original/refined timestamps: %+v", parsed.Spots[0])
	}

	withoutPresenter, err := os.ReadFile(filepath.Join("testdata", "response_without_presenter.json"))
	if err != nil {
		t.Fatalf("read response_without_presenter fixture: %v", err)
	}
	parsed, err = ParseAndValidateResponse(withoutPresenter)
	if err != nil {
		t.Fatalf("parse response_without_presenter fixture: %v", err)
	}
	if parsed.PresenterName != nil {
		t.Fatalf("expected nil presenter_name for fixture without presenter, got %+v", parsed.PresenterName)
	}
}
