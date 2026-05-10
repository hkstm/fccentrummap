package articletext

import (
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func TestExtractArticleTextContentTrafilaturaMatched(t *testing.T) {
	html := `<html><body><nav>Menu</nav><article><h1>Title</h1><p>This is a longer first paragraph with enough descriptive text to pass extraction threshold.</p><p>This is a longer second paragraph that keeps the article body clearly content-heavy.</p></article></body></html>`

	result := ExtractArticleTextContent(html)

	if result.Status != models.ArticleTextExtractionStatusMatched {
		t.Fatalf("status mismatch: got %s", result.Status)
	}
	if result.ExtractionMode != models.ArticleTextExtractionModeTrafilatura {
		t.Fatalf("mode mismatch: got %s", result.ExtractionMode)
	}
	if result.MatchedCount == 0 || len(result.Contents) == 0 {
		t.Fatalf("expected matched content, got count=%d len=%d", result.MatchedCount, len(result.Contents))
	}
	if result.Contents[0].SourceType != models.ArticleTextSourceTypeTrafilaturaText {
		t.Fatalf("source type mismatch: got %s", result.Contents[0].SourceType)
	}
}

func TestExtractArticleTextContentNoMatch(t *testing.T) {
	html := `<html><head><title>Empty</title></head><body><script>var x=1;</script></body></html>`

	result := ExtractArticleTextContent(html)

	if result.Status != models.ArticleTextExtractionStatusNoMatch {
		t.Fatalf("status mismatch: got %s", result.Status)
	}
	if result.ExtractionMode != models.ArticleTextExtractionModeNoMatch {
		t.Fatalf("mode mismatch: got %s", result.ExtractionMode)
	}
	if result.MatchedCount != 0 {
		t.Fatalf("matched count mismatch: got %d", result.MatchedCount)
	}
	if len(result.Contents) != 0 {
		t.Fatalf("expected no contents, got %d", len(result.Contents))
	}
}

func TestExtractArticleTextContentParserError(t *testing.T) {
	html := "   "

	result := ExtractArticleTextContent(html)

	if result.Status != models.ArticleTextExtractionStatusError {
		t.Fatalf("status mismatch: got %s", result.Status)
	}
	if result.ExtractionMode != models.ArticleTextExtractionModeError {
		t.Fatalf("mode mismatch: got %s", result.ExtractionMode)
	}
	if result.ErrorMessage == nil || !strings.Contains(*result.ErrorMessage, "empty html") {
		t.Fatalf("expected empty html error message, got %v", result.ErrorMessage)
	}
}

func TestNormalizeTrafilaturaText(t *testing.T) {
	out := normalizeTrafilaturaText("  One\n\nTwo  \n   \nThree")
	if len(out) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(out))
	}
	if out[0].Content != "One" || out[2].Content != "Three" {
		t.Fatalf("unexpected content: %+v", out)
	}
	if out[0].SourceType != models.ArticleTextSourceTypeTrafilaturaText {
		t.Fatalf("unexpected source type: %s", out[0].SourceType)
	}
}
