package scraper

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hkstm/fccentrummap/internal/repository"
)

type fakeDownloader struct {
	calls int
	fn    func(videoID string) (*DownloadedAudio, error)
}

func (f *fakeDownloader) Download(_ context.Context, videoID string) (*DownloadedAudio, error) {
	f.calls++
	return f.fn(videoID)
}

func tempRepo(t *testing.T) *repository.Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return repo
}

func writeTempAudioFile(t *testing.T, ext string, content []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "audio."+ext)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write temp audio: %v", err)
	}
	return path
}

func TestAcquireAndStoreAudioStoresBlobAndMetadata(t *testing.T) {
	repo := tempRepo(t)
	videoID := "dQw4w9WgXcQ"
	if err := repo.InsertArticleRaw("https://example.com/a", "<html></html>", &videoID); err != nil {
		t.Fatalf("insert article: %v", err)
	}

	audioPath := writeTempAudioFile(t, "mp3", []byte("audio-bytes"))
	downloader := &fakeDownloader{fn: func(_ string) (*DownloadedAudio, error) {
		return &DownloadedAudio{
			Path:       audioPath,
			Format:     "mp3",
			MIMEType:   "audio/mpeg",
			YouTubeURL: "https://www.youtube.com/watch?v=" + videoID,
		}, nil
	}}

	if err := AcquireAndStoreAudio(context.Background(), repo, downloader); err != nil {
		t.Fatalf("acquire and store audio: %v", err)
	}

	pending, err := repo.GetPendingArticles()
	if err != nil {
		t.Fatalf("get pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending article, got %d", len(pending))
	}

	src, err := repo.GetArticleAudioSource(pending[0].ArticleRawID)
	if err != nil {
		t.Fatalf("get audio source: %v", err)
	}
	if src == nil {
		t.Fatal("expected audio source row")
	}
	if src.AudioFormat != "mp3" {
		t.Fatalf("audio format mismatch: got %s", src.AudioFormat)
	}
	if src.MIMEType != "audio/mpeg" {
		t.Fatalf("mime type mismatch: got %s", src.MIMEType)
	}
	if src.ByteSize != int64(len("audio-bytes")) {
		t.Fatalf("byte size mismatch: got %d", src.ByteSize)
	}
}

func TestAcquireAndStoreAudioSkipsExistingRowsForIdempotency(t *testing.T) {
	repo := tempRepo(t)
	videoID := "dQw4w9WgXcQ"
	if err := repo.InsertArticleRaw("https://example.com/b", "<html></html>", &videoID); err != nil {
		t.Fatalf("insert article: %v", err)
	}

	audioPath := writeTempAudioFile(t, "wav", []byte("wav-bytes"))
	downloader := &fakeDownloader{fn: func(_ string) (*DownloadedAudio, error) {
		return &DownloadedAudio{
			Path:       audioPath,
			Format:     "wav",
			MIMEType:   "audio/wav",
			YouTubeURL: "https://www.youtube.com/watch?v=" + videoID,
		}, nil
	}}

	if err := AcquireAndStoreAudio(context.Background(), repo, downloader); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := AcquireAndStoreAudio(context.Background(), repo, downloader); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if downloader.calls != 1 {
		t.Fatalf("expected downloader to be called once, got %d", downloader.calls)
	}
}

func TestAcquireAndStoreAudioFailureIsRetryable(t *testing.T) {
	repo := tempRepo(t)
	videoID := "dQw4w9WgXcQ"
	if err := repo.InsertArticleRaw("https://example.com/c", "<html></html>", &videoID); err != nil {
		t.Fatalf("insert article: %v", err)
	}

	first := &fakeDownloader{fn: func(_ string) (*DownloadedAudio, error) {
		return nil, errors.New("yt-dlp failed")
	}}
	err := AcquireAndStoreAudio(context.Background(), repo, first)
	if err == nil {
		t.Fatal("expected error on failed download")
	}

	pending, err := repo.GetPendingArticles()
	if err != nil {
		t.Fatalf("get pending: %v", err)
	}
	src, err := repo.GetArticleAudioSource(pending[0].ArticleRawID)
	if err != nil {
		t.Fatalf("get audio source: %v", err)
	}
	if src != nil {
		t.Fatal("expected no audio source after failed download")
	}

	audioPath := writeTempAudioFile(t, "m4a", []byte("retry-bytes"))
	second := &fakeDownloader{fn: func(_ string) (*DownloadedAudio, error) {
		return &DownloadedAudio{
			Path:       audioPath,
			Format:     "m4a",
			MIMEType:   "audio/mp4",
			YouTubeURL: "https://www.youtube.com/watch?v=" + videoID,
		}, nil
	}}
	if err := AcquireAndStoreAudio(context.Background(), repo, second); err != nil {
		t.Fatalf("retry run failed: %v", err)
	}

	src, err = repo.GetArticleAudioSource(pending[0].ArticleRawID)
	if err != nil {
		t.Fatalf("get audio source after retry: %v", err)
	}
	if src == nil {
		t.Fatal("expected audio source row after successful retry")
	}
	if src.AudioFormat != "m4a" {
		t.Fatalf("audio format mismatch after retry: got %s", src.AudioFormat)
	}
	if src.MIMEType != "audio/mp4" {
		t.Fatalf("mime type mismatch after retry: got %s", src.MIMEType)
	}
	if src.ByteSize != int64(len("retry-bytes")) {
		t.Fatalf("byte size mismatch after retry: got %d", src.ByteSize)
	}
}
