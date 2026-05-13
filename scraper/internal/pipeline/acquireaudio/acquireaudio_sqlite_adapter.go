package acquireaudio

import (
	"context"

	"github.com/hkstm/fccentrummap/internal/audio"
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
	if err := audio.AcquireAndStoreAudio(ctx, repo, nil); err != nil {
		return Response{}, err
	}
	return Response{Identity: "acquire-audio", Stage: "acquireaudio"}, nil
}
