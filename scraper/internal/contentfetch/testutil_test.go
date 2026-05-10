package contentfetch

import (
	"path/filepath"
	"testing"

	"github.com/hkstm/fccentrummap/internal/repository"
)

func tempRepo(t *testing.T) *repository.Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "spots.db")
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("repository.New: %v", err)
	}
	if err := repo.InitSchema(); err != nil {
		repo.Close()
		t.Fatalf("repo.InitSchema: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return repo
}
