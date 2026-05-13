package exportdata

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(_ context.Context, req Request) (Response, error) {
	repo, err := repository.New(req.DBPath)
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()

	data, err := repo.ExportData()
	if err != nil {
		return Response{}, err
	}
	outPath := req.OutputPath
	if outPath == "" {
		outPath = filepath.Clean("../viz/public/data/spots.json")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return Response{}, err
	}
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return Response{}, err
	}
	if err := os.WriteFile(outPath, payload, 0o644); err != nil {
		return Response{}, err
	}
	return Response{Identity: "export-data", Stage: "exportdata", OutputPath: outPath}, nil
}
