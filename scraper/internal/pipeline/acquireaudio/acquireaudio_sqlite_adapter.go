package acquireaudio

import (
	"context"
	"strings"

	"github.com/hkstm/fccentrummap/internal/audio"
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
	if err := audio.AcquireAndStoreAudio(ctx, repo, nil); err != nil {
		return Response{}, err
	}
	return Response{Identity: req.Identity, Stage: "acquireaudio"}, nil
}
