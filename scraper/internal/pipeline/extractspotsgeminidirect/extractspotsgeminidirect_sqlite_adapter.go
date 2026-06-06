package extractspotsgeminidirect

import (
	"context"

	"github.com/hkstm/fccentrummap/internal/geminidirect"
	genaiclient "github.com/hkstm/fccentrummap/internal/genai"
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

	client := genaiclient.NewClientWithEndpoint(req.APIKey, req.Model, req.Endpoint)
	runner := Runner{
		Inputs:    repositoryInputLoader{repo: repo},
		Client:    client,
		Writer:    geminidirect.ArtifactWriter{OutDir: req.OutDir},
		Persister: repositoryAcceptedSpotPersister{repo: repo},
		Model:     req.Model,
		Force:     req.Force,
		Limit:     req.Limit,
	}
	return runner.Run(ctx)
}

type repositoryInputLoader struct {
	repo *repository.Repository
}

type repositoryAcceptedSpotPersister struct {
	repo *repository.Repository
}

func (p repositoryAcceptedSpotPersister) PersistAcceptedGeminiDirectArticle(article AcceptedArticle) error {
	if article.PresenterName != nil {
		presenterName := *article.PresenterName
		if presenterName != "" {
			presenterID, err := p.repo.UpsertPresenter(presenterName)
			if err != nil {
				return err
			}
			if err := p.repo.LinkArticlePresenter(article.ArticleSourceID, presenterID); err != nil {
				return err
			}
		}
	}
	for _, spot := range article.Spots {
		if _, err := p.repo.UpsertGeminiDirectSpotMention(repository.GeminiDirectSpotMentionInput{
			ArticleSourceID:         spot.ArticleSourceID,
			ArticleURL:              spot.ArticleURL,
			YouTubeURL:              spot.YouTubeURL,
			Place:                   spot.Place,
			Address:                 spot.Address,
			YouTubeTimestampSeconds: spot.YouTubeTimestampSeconds,
			Evidence:                spot.Evidence,
			Confidence:              spot.Confidence,
			Model:                   spot.Model,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (l repositoryInputLoader) ListGeminiDirectInputs() ([]ArticleInput, error) {
	rows, err := l.repo.ListGeminiDirectInputs()
	if err != nil {
		return nil, err
	}
	out := make([]ArticleInput, 0, len(rows))
	for _, row := range rows {
		out = append(out, ArticleInput{ArticleSourceID: row.ArticleSourceID, ArticleURL: row.ArticleURL, YouTubeURL: row.YouTubeURL})
	}
	return out, nil
}
