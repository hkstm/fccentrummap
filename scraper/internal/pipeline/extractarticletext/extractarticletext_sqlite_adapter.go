package extractarticletext

import (
	"context"
	"strings"

	"github.com/hkstm/fccentrummap/internal/articletext"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(ctx context.Context, req Request) (Response, error) {
	repo, err := repository.New(req.DBPath)
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}

	fetches, err := repo.ListArticleFetches()
	if err != nil {
		return Response{}, err
	}
	processed := 0
	for _, f := range fetches {
		select {
		case <-ctx.Done():
			return Response{}, ctx.Err()
		default:
		}
		extraction := articletext.ExtractArticleTextContent(f.HTML)
		parts := make([]string, 0, len(extraction.Contents))
		for _, c := range extraction.Contents {
			if trimmed := strings.TrimSpace(c.Content); trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
		cleaned := strings.TrimSpace(strings.Join(parts, "\n\n"))
		if cleaned == "" {
			continue
		}
		if _, err := repo.UpsertArticleText(f.ArticleFetchID, cleaned); err != nil {
			return Response{}, err
		}
		processed++
	}

	return Response{Identity: "extract-article-text", Stage: "extractarticletext", ProcessedCount: processed}, nil
}
