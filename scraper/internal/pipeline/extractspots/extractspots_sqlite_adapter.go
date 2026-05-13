package extractspots

import (
	"context"
	"fmt"

	"github.com/hkstm/fccentrummap/internal/extractspots"
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

	res, err := extractspots.Run(ctx, repo, extractspots.Options{
		OutDir:     req.OutDir,
		GemmaModel: req.GemmaModel,
		APIKey:     req.APIKey,
		Endpoint:   req.Endpoint,
	})
	if err != nil {
		return Response{}, err
	}

	return Response{Identity: fmt.Sprintf("spot-extraction-%d", res.SpotExtractionID), Stage: "extractspots", SpotExtractionID: res.SpotExtractionID}, nil
}
