package collectarticleurls

import (
	"context"

	"github.com/hkstm/fccentrummap/internal/contentfetch"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(_ context.Context, req Request) (Response, error) {
	repo, err := repository.New(req.DBPath)
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}

	if u := req.ArticleURL; u != "" {
		if _, err := repo.UpsertArticleSource(u); err != nil {
			return Response{}, err
		}
		return Response{Identity: u, Stage: "collectarticleurls", URLs: []string{u}}, nil
	}

	urls, err := contentfetch.CrawlArticleURLs()
	if err != nil {
		return Response{}, err
	}
	for _, u := range urls {
		if _, err := repo.UpsertArticleSource(u); err != nil {
			return Response{}, err
		}
	}
	return Response{Identity: "collect-article-urls", Stage: "collectarticleurls", URLs: urls}, nil
}
