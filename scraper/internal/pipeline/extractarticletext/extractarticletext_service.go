package extractarticletext

import (
	"context"
	"fmt"
	"strings"
)

type Mode string

const (
	ModeSQLite Mode = "sqlite"
	ModeFile   Mode = "file"
)

func parseMode(mode string) (Mode, error) {
	normalized := Mode(strings.ToLower(strings.TrimSpace(mode)))
	switch normalized {
	case ModeSQLite, ModeFile:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported mode: %s", mode)
	}
}

type Service struct {
	sqlite SQLitePort
	file   FilePort
}

func NewService(sqlite SQLitePort, file FilePort) *Service {
	return &Service{sqlite: sqlite, file: file}
}

func (s *Service) Run(ctx context.Context, mode string, req Request) (Response, error) {
	parsedMode, err := parseMode(mode)
	if err != nil {
		return Response{}, err
	}

	switch parsedMode {
	case ModeSQLite:
		if s.sqlite == nil {
			return Response{}, fmt.Errorf("sqlite adapter not configured")
		}
		return s.sqlite.Run(ctx, req)
	case ModeFile:
		if s.file == nil {
			return Response{}, fmt.Errorf("file adapter not configured")
		}
		return s.file.Run(ctx, req)
	default:
		return Response{}, fmt.Errorf("unsupported mode: %s", mode)
	}
}
