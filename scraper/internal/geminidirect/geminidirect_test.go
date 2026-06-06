package geminidirect

import (
	"strings"
	"testing"
)

func TestBuildPromptUsesOnlyURLsAsPrimaryInputs(t *testing.T) {
	prompt, err := BuildPrompt(PromptInput{ArticleURL: "https://example.test/article", YouTubeURL: "https://youtube.com/watch?v=abc"})
	if err != nil {
		t.Fatalf("BuildPrompt: %v", err)
	}
	for _, want := range []string{"article_url: https://example.test/article", "youtube_url: https://youtube.com/watch?v=abc", "Use the article_url and youtube_url below as the primary source inputs"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "[cleaned_article]") {
		t.Fatalf("prompt should not contain cleaned article text sections: %s", prompt)
	}
}

func TestGenerateContentConfigDeclaresStructuredJSONResponse(t *testing.T) {
	cfg := GenerateContentConfig()
	if cfg == nil {
		t.Fatalf("expected config")
	}
	if cfg.ResponseMIMEType != "application/json" {
		t.Fatalf("response MIME type = %q", cfg.ResponseMIMEType)
	}
	if cfg.ResponseSchema == nil {
		t.Fatalf("expected response schema")
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0].URLContext == nil {
		t.Fatalf("Gemini-direct should enable URL context")
	}
	if len(cfg.Tools[0].FunctionDeclarations) != 0 || cfg.ToolConfig != nil {
		t.Fatalf("Gemini-direct should use native structured JSON, not function calling")
	}
	spotProps := cfg.ResponseSchema.Properties["spots"].Items.Properties
	for _, field := range []string{"place", "youtubeTimestampSeconds", "evidence", "confidence", "notes"} {
		if spotProps[field] == nil {
			t.Fatalf("missing spot field schema %s", field)
		}
	}
}

func TestParseAndValidateResponse(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"parts":[{"text":"{\"presenter_name\":\"  Jane  \",\"spots\":[{\"place\":\"  Cafe Binnenvisser \",\"youtubeTimestampSeconds\":42,\"evidence\":\"shown in video\",\"confidence\":0.9,\"notes\":\"review\"}]}"}]}}]}`)
	parsed, err := ParseAndValidateResponse(raw)
	if err != nil {
		t.Fatalf("ParseAndValidateResponse: %v", err)
	}
	if parsed.PresenterName == nil || *parsed.PresenterName != "Jane" {
		t.Fatalf("presenter = %#v", parsed.PresenterName)
	}
	if got := parsed.Spots[0].Place; got != "Cafe Binnenvisser" {
		t.Fatalf("place = %q", got)
	}
}

func TestParseAndValidateResponseMalformedDiagnostics(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"parts":[{"text":"{\"spots\":[{\"place\":\"\",\"youtubeTimestampSeconds\":1,\"evidence\":\"x\",\"confidence\":0.5,\"notes\":\"\"}]}"}]}}]}`)
	_, err := ParseAndValidateResponse(raw)
	if err == nil || !strings.Contains(err.Error(), "spots[0].place is required") {
		t.Fatalf("expected place diagnostic, got %v", err)
	}
}
