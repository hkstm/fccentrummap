package extractspots

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/extractspots"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(ctx context.Context, req Request) (Response, error) {
	repo, err := repository.New(strings.TrimSpace(req.DBPath))
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}

	res, err := extractspots.Run(ctx, repo, extractspots.Options{
		UseLatest:     req.UseLatest,
		OutDir:        strings.TrimSpace(req.OutDir),
		GemmaModel:    strings.TrimSpace(req.GemmaModel),
		APIKey:        strings.TrimSpace(req.APIKey),
		Endpoint:      strings.TrimSpace(req.Endpoint),
		PersistRecord: true,
	})
	if err != nil {
		return Response{}, err
	}

	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = fmt.Sprintf("spot-extraction-%d", res.SpotExtractionID)
	}
	return Response{Identity: identity, Stage: "extractspots", SpotExtractionID: res.SpotExtractionID}, nil
}
