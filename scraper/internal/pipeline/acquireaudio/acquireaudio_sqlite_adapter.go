package acquireaudio

import (
	"context"
	"fmt"
	"os"

	"github.com/hkstm/fccentrummap/internal/audio"
	"github.com/hkstm/fccentrummap/internal/contentfetch"
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
	d := &audio.YTDLPDownloader{}
	for _, f := range fetches {
		videoID, ok := contentfetch.ExtractYouTubeVideoID(f.HTML)
		if !ok {
			continue
		}
		downloaded, err := d.Download(ctx, videoID)
		if err != nil {
			return Response{}, fmt.Errorf("download audio for article_fetch_id=%d: %w", f.ArticleFetchID, err)
		}
		blob, err := os.ReadFile(downloaded.Path)
		if err != nil {
			return Response{}, fmt.Errorf("read downloaded audio for article_fetch_id=%d: %w", f.ArticleFetchID, err)
		}
		_ = os.Remove(downloaded.Path)
		if _, err := repo.UpsertAudioSource(f.ArticleFetchID, downloaded.YouTubeURL, downloaded.Format, downloaded.MIMEType, blob); err != nil {
			return Response{}, err
		}
	}
	return Response{Identity: "acquire-audio", Stage: "acquireaudio"}, nil
}
