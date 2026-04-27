package scraper

import (
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func TestArticleTextExtractionPersistencePipelineWithRepresentativeOutcomes(t *testing.T) {
	repo := tempRepo(t)

	fixtures := []struct {
		url                string
		html               string
		expectedStatus     string
		expectedMode       string
		expectedContentLen int
	}{
		{
			url:                "https://example.com/matched",
			html:               `<html><body><article><p>This matched article has enough body text to clear the minimal extraction threshold and should be persisted.</p></article></body></html>`,
			expectedStatus:     models.ArticleTextExtractionStatusMatched,
			expectedMode:       models.ArticleTextExtractionModeTrafilatura,
			expectedContentLen: 1,
		},
		{
			url:                "https://example.com/no-match",
			html:               `<html><head><title>No body text</title></head><body><script>var x=1;</script></body></html>`,
			expectedStatus:     models.ArticleTextExtractionStatusNoMatch,
			expectedMode:       models.ArticleTextExtractionModeNoMatch,
			expectedContentLen: 0,
		},
		{
			url:                "https://example.com/error",
			html:               ``,
			expectedStatus:     models.ArticleTextExtractionStatusError,
			expectedMode:       models.ArticleTextExtractionModeError,
			expectedContentLen: 0,
		},
	}

	for _, fixture := range fixtures {
		if err := repo.InsertArticleRaw(fixture.url, fixture.html, nil); err != nil {
			t.Fatalf("insert article_raw %s: %v", fixture.url, err)
		}

		articleRaw, err := repo.GetArticleRawByURL(fixture.url)
		if err != nil {
			t.Fatalf("get article_raw by url %s: %v", fixture.url, err)
		}
		if articleRaw == nil {
			t.Fatalf("expected article_raw row for %s", fixture.url)
		}

		result := ExtractArticleTextContent(articleRaw.HTML)
		result.ArticleRawID = articleRaw.ArticleRawID
		if err := repo.ReplaceArticleTextExtraction(result); err != nil {
			t.Fatalf("replace extraction %s: %v", fixture.url, err)
		}
	}

	for _, fixture := range fixtures {
		articleRaw, err := repo.GetArticleRawByURL(fixture.url)
		if err != nil {
			t.Fatalf("reload article_raw by url %s: %v", fixture.url, err)
		}
		if articleRaw == nil {
			t.Fatalf("missing article_raw row for %s", fixture.url)
		}

		extraction, err := repo.GetArticleTextExtraction(articleRaw.ArticleRawID)
		if err != nil {
			t.Fatalf("get extraction for %s: %v", fixture.url, err)
		}
		if extraction == nil {
			t.Fatalf("expected extraction row for %s", fixture.url)
		}
		if extraction.Status != fixture.expectedStatus {
			t.Fatalf("status mismatch for %s: got %s want %s", fixture.url, extraction.Status, fixture.expectedStatus)
		}
		if extraction.ExtractionMode != fixture.expectedMode {
			t.Fatalf("mode mismatch for %s: got %s want %s", fixture.url, extraction.ExtractionMode, fixture.expectedMode)
		}

		contents, err := repo.ListArticleTextContents(articleRaw.ArticleRawID)
		if err != nil {
			t.Fatalf("list contents for %s: %v", fixture.url, err)
		}
		if len(contents) != fixture.expectedContentLen {
			t.Fatalf("content count mismatch for %s: got %d want %d", fixture.url, len(contents), fixture.expectedContentLen)
		}
	}
}
