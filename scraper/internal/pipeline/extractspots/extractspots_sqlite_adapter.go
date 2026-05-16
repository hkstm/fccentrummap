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

	results, err := extractspots.Run(ctx, repo, extractspots.Options{
		OutDir:     req.OutDir,
		Model: req.Model,
		APIKey:     req.APIKey,
		Endpoint:   req.Endpoint,
	})
	if err != nil {
		return Response{}, err
	}

	var processedIDs []int64
	for _, res := range results {
		processedIDs = append(processedIDs, res.SpotExtractionID)
	}

	if len(processedIDs) == 0 {
		return Response{Identity: "spot-extraction-none", Stage: "extractspots"}, nil
	}

	return Response{
		Identity:          fmt.Sprintf("spot-extraction-batch-%d", processedIDs[len(processedIDs)-1]),
		Stage:             "extractspots",
		SpotExtractionIDs: processedIDs,
	}, nil
}
