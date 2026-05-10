package articletext

import (
	"strings"
	"unicode/utf8"

	"github.com/hkstm/fccentrummap/internal/models"
	trafilatura "github.com/markusmobius/go-trafilatura"
)

const minTrafilaturaTextChars = 80

func ExtractArticleTextContent(html string) models.ArticleTextExtractionResult {
	if strings.TrimSpace(html) == "" {
		errMsg := "empty html input"
		return models.ArticleTextExtractionResult{
			ExtractionMode: models.ArticleTextExtractionModeError,
			Status:         models.ArticleTextExtractionStatusError,
			MatchedCount:   0,
			ErrorMessage:   &errMsg,
		}
	}

	res, err := trafilatura.Extract(strings.NewReader(html), trafilatura.Options{EnableFallback: true})
	if err != nil {
		errMsg := err.Error()
		return models.ArticleTextExtractionResult{
			ExtractionMode: models.ArticleTextExtractionModeError,
			Status:         models.ArticleTextExtractionStatusError,
			MatchedCount:   0,
			ErrorMessage:   &errMsg,
		}
	}

	contents := normalizeTrafilaturaText(res.ContentText)
	totalChars := 0
	for _, c := range contents {
		totalChars += utf8.RuneCountInString(c.Content)
	}
	if len(contents) == 0 || totalChars < minTrafilaturaTextChars {
		return models.ArticleTextExtractionResult{
			ExtractionMode: models.ArticleTextExtractionModeNoMatch,
			Status:         models.ArticleTextExtractionStatusNoMatch,
			MatchedCount:   0,
			Contents:       nil,
		}
	}

	return models.ArticleTextExtractionResult{
		ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
		Status:         models.ArticleTextExtractionStatusMatched,
		MatchedCount:   len(contents),
		Contents:       contents,
	}
}

func normalizeTrafilaturaText(text string) []models.ArticleTextContentInput {
	parts := strings.Split(text, "\n")
	results := make([]models.ArticleTextContentInput, 0, len(parts))
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" {
			continue
		}
		results = append(results, models.ArticleTextContentInput{
			SourceType: models.ArticleTextSourceTypeTrafilaturaText,
			Content:    line,
		})
	}
	return results
}
