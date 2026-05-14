package repository

import (
	"path/filepath"
	"testing"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return repo
}
