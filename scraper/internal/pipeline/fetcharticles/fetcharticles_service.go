package fetcharticles

import (
	"context"
	"fmt"
)

type Service struct {
	sqlite SQLitePort
	file   FilePort
}

func NewService(sqlite SQLitePort, file FilePort) *Service {
	return &Service{sqlite: sqlite, file: file}
}

func (s *Service) Run(ctx context.Context, mode string, req Request) (Response, error) {
	switch mode {
	case "sqlite":
		if s.sqlite == nil {
			return Response{}, fmt.Errorf("sqlite adapter not configured")
		}
		return s.sqlite.Run(ctx, req)
	case "file":
		if s.file == nil {
			return Response{}, fmt.Errorf("file adapter not configured")
		}
		return s.file.Run(ctx, req)
	default:
		return Response{}, fmt.Errorf("unsupported mode: %s", mode)
	}
}
