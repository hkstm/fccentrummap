package repository

import (
	"path/filepath"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return repo
}

func insertTestArticleRaw(t *testing.T, repo *Repository, url string) int64 {
	t.Helper()
	if err := repo.InsertArticleRaw(url, "<html></html>", nil); err != nil {
		t.Fatalf("insert article_raw: %v", err)
	}
	row, err := repo.GetArticleRawByURL(url)
	if err != nil {
		t.Fatalf("get article_raw by url: %v", err)
	}
	if row == nil {
		t.Fatal("expected article_raw row")
	}
	return row.ArticleRawID
}

func TestInitSchemaIsIdempotentAndPreservesExtractionData(t *testing.T) {
	repo := newTestRepo(t)
	articleRawID := insertTestArticleRaw(t, repo, "https://example.com/idempotent")

	err := repo.ReplaceArticleTextExtraction(models.ArticleTextExtractionResult{
		ArticleRawID:   articleRawID,
		ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
		Status:         models.ArticleTextExtractionStatusMatched,
		MatchedCount:   1,
		Contents: []models.ArticleTextContentInput{{
			SourceType: models.ArticleTextSourceTypeTrafilaturaText,
			Content:    "hello",
		}},
	})
	if err != nil {
		t.Fatalf("replace article text extraction: %v", err)
	}

	if err := repo.InitSchema(); err != nil {
		t.Fatalf("second init schema failed: %v", err)
	}

	extraction, err := repo.GetArticleTextExtraction(articleRawID)
	if err != nil {
		t.Fatalf("get extraction: %v", err)
	}
	if extraction == nil {
		t.Fatal("expected extraction row to remain after re-init")
	}
	contents, err := repo.ListArticleTextContents(articleRawID)
	if err != nil {
		t.Fatalf("list contents: %v", err)
	}
	if len(contents) != 1 {
		t.Fatalf("expected 1 content row, got %d", len(contents))
	}
}

func TestReplaceArticleTextExtractionAtomicRollbackOnConstraintFailure(t *testing.T) {
	repo := newTestRepo(t)
	articleRawID := insertTestArticleRaw(t, repo, "https://example.com/atomic")

	err := repo.ReplaceArticleTextExtraction(models.ArticleTextExtractionResult{
		ArticleRawID:   articleRawID,
		ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
		Status:         models.ArticleTextExtractionStatusMatched,
		MatchedCount:   1,
		Contents: []models.ArticleTextContentInput{{
			SourceType: models.ArticleTextSourceTypeTrafilaturaText,
			Content:    "original",
		}},
	})
	if err != nil {
		t.Fatalf("seed replace failed: %v", err)
	}

	err = repo.ReplaceArticleTextExtraction(models.ArticleTextExtractionResult{
		ArticleRawID:   articleRawID,
		ExtractionMode: models.ArticleTextExtractionModeError,
		Status:         "invalid-status",
		MatchedCount:   0,
	})
	if err == nil {
		t.Fatal("expected replace failure for invalid status")
	}

	extraction, err := repo.GetArticleTextExtraction(articleRawID)
	if err != nil {
		t.Fatalf("get extraction after rollback: %v", err)
	}
	if extraction == nil {
		t.Fatal("expected extraction row after rollback")
	}
	if extraction.Status != models.ArticleTextExtractionStatusMatched {
		t.Fatalf("expected matched status to remain, got %s", extraction.Status)
	}
	contents, err := repo.ListArticleTextContents(articleRawID)
	if err != nil {
		t.Fatalf("list contents after rollback: %v", err)
	}
	if len(contents) != 1 || contents[0].Content != "original" {
		t.Fatalf("expected original content to remain after rollback, got %+v", contents)
	}
}

func TestReplaceArticleTextExtractionRejectsInvalidPayloadInvariants(t *testing.T) {
	repo := newTestRepo(t)
	articleRawID := insertTestArticleRaw(t, repo, "https://example.com/invariants")

	cases := []models.ArticleTextExtractionResult{
		{
			ArticleRawID:   articleRawID,
			ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
			Status:         models.ArticleTextExtractionStatusMatched,
			MatchedCount:   0,
			Contents: []models.ArticleTextContentInput{
				{SourceType: models.ArticleTextSourceTypeTrafilaturaText, Content: "one"},
			},
		},
		{
			ArticleRawID:   articleRawID,
			ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
			Status:         models.ArticleTextExtractionStatusMatched,
			MatchedCount:   2,
			Contents: []models.ArticleTextContentInput{
				{SourceType: models.ArticleTextSourceTypeTrafilaturaText, Content: "one"},
			},
		},
		{
			ArticleRawID:   articleRawID,
			ExtractionMode: models.ArticleTextExtractionModeNoMatch,
			Status:         models.ArticleTextExtractionStatusNoMatch,
			MatchedCount:   1,
		},
		{
			ArticleRawID:   articleRawID,
			ExtractionMode: models.ArticleTextExtractionModeError,
			Status:         models.ArticleTextExtractionStatusError,
			MatchedCount:   0,
			Contents: []models.ArticleTextContentInput{
				{SourceType: models.ArticleTextSourceTypeTrafilaturaText, Content: "invalid"},
			},
		},
	}

	for i, tc := range cases {
		if err := repo.ReplaceArticleTextExtraction(tc); err == nil {
			t.Fatalf("expected invariant validation error for case %d", i)
		}
	}
}

func TestReplaceArticleTextExtractionIdempotentRerunSemantics(t *testing.T) {
	repo := newTestRepo(t)
	articleRawID := insertTestArticleRaw(t, repo, "https://example.com/rerun")

	matched := models.ArticleTextExtractionResult{
		ArticleRawID:   articleRawID,
		ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
		Status:         models.ArticleTextExtractionStatusMatched,
		MatchedCount:   2,
		Contents: []models.ArticleTextContentInput{
			{SourceType: models.ArticleTextSourceTypeTrafilaturaText, Content: "one"},
			{SourceType: models.ArticleTextSourceTypeTrafilaturaText, Content: "two"},
		},
	}
	if err := repo.ReplaceArticleTextExtraction(matched); err != nil {
		t.Fatalf("first replace failed: %v", err)
	}
	if err := repo.ReplaceArticleTextExtraction(matched); err != nil {
		t.Fatalf("second replace failed: %v", err)
	}

	extraction, err := repo.GetArticleTextExtraction(articleRawID)
	if err != nil {
		t.Fatalf("get extraction after rerun: %v", err)
	}
	if extraction == nil {
		t.Fatal("expected extraction row after rerun")
	}
	if extraction.MatchedCount != 2 {
		t.Fatalf("matched_count mismatch: got %d", extraction.MatchedCount)
	}
	contents, err := repo.ListArticleTextContents(articleRawID)
	if err != nil {
		t.Fatalf("list contents after rerun: %v", err)
	}
	if len(contents) != 2 {
		t.Fatalf("expected 2 content rows after rerun, got %d", len(contents))
	}

	if err := repo.ReplaceArticleTextExtraction(models.ArticleTextExtractionResult{
		ArticleRawID:   articleRawID,
		ExtractionMode: models.ArticleTextExtractionModeNoMatch,
		Status:         models.ArticleTextExtractionStatusNoMatch,
		MatchedCount:   0,
	}); err != nil {
		t.Fatalf("replace with no_match failed: %v", err)
	}

	extraction, err = repo.GetArticleTextExtraction(articleRawID)
	if err != nil {
		t.Fatalf("get extraction after no_match: %v", err)
	}
	if extraction == nil || extraction.Status != models.ArticleTextExtractionStatusNoMatch {
		t.Fatalf("expected no_match extraction, got %+v", extraction)
	}
	contents, err = repo.ListArticleTextContents(articleRawID)
	if err != nil {
		t.Fatalf("list contents after no_match: %v", err)
	}
	if len(contents) != 0 {
		t.Fatalf("expected 0 content rows for no_match, got %d", len(contents))
	}
}
