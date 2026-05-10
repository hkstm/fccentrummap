package exportdata

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type inputPayload struct {
	Authors []any `json:"authors"`
	Spots   []any `json:"spots"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, req Request) (Response, error) {
	if strings.TrimSpace(req.InputPath) == "" {
		return Response{}, fmt.Errorf("exportdata file input requires inputPath")
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = common.IdentityFromPath(req.InputPath)
	}
	in, err := common.ReadJSON[inputPayload](req.InputPath)
	if err != nil {
		return Response{}, err
	}
	if len(in.Authors) == 0 && len(in.Spots) == 0 {
		return Response{}, fmt.Errorf("exportdata input requires authors or spots")
	}
	out, err := common.WriteStageArtifact("", "export-data", identity, "export", map[string]any{"identity": identity, "stage": "exportdata", "authors": in.Authors, "spots": in.Spots})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "exportdata", OutputPath: out}, nil
}
