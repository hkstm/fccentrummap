package fetcharticles

import (
	"context"
	"strings"

	"github.com/hkstm/fccentrummap/internal/contentfetch"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(_ context.Context, req Request) (Response, error) {
	repo, err := repository.New(strings.TrimSpace(req.DBPath))
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}
	articles, err := repo.GetPendingArticles()
	if err != nil {
		return Response{}, err
	}
	urls := make([]string, 0, len(articles))
	for _, a := range articles {
		urls = append(urls, a.URL)
	}
	if err := contentfetch.FetchAndStoreArticles(urls, repo); err != nil {
		return Response{}, err
	}
	return Response{Identity: req.Identity, Stage: "fetcharticles", ArticleURLs: urls, FetchedCount: len(urls)}, nil
}
