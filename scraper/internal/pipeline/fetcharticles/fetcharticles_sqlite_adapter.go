package fetcharticles

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

	sources, err := repo.ListArticleSources()
	if err != nil {
		return Response{}, err
	}
	urls := make([]string, 0, len(sources))
	for _, s := range sources {
		urls = append(urls, s.URL)
	}
	htmlByURL, err := contentfetch.FetchArticlesHTML(urls)
	if err != nil {
		return Response{}, err
	}
	for _, s := range sources {
		html, ok := htmlByURL[s.URL]
		if !ok {
			continue
		}
		if _, err := repo.UpsertArticleFetch(s.ArticleSourceID, html); err != nil {
			return Response{}, err
		}
	}
	return Response{Identity: "fetch-articles", Stage: "fetcharticles", ArticleURLs: urls, FetchedCount: len(htmlByURL)}, nil
}
