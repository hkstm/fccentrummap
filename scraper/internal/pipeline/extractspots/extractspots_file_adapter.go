package extractspots

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type inputPayload struct {
	TranscriptionJSON string `json:"transcriptionJson"`
	ArticleText       string `json:"articleText"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, req Request) (Response, error) {
	if strings.TrimSpace(req.InputPath) == "" {
		return Response{}, fmt.Errorf("extractspots file input requires inputPath")
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = common.IdentityFromPath(req.InputPath)
	}
	in, err := common.ReadJSON[inputPayload](req.InputPath)
	if err != nil {
		return Response{}, err
	}
	if strings.TrimSpace(in.TranscriptionJSON) == "" || strings.TrimSpace(in.ArticleText) == "" {
		return Response{}, fmt.Errorf("extractspots input requires transcriptionJson and articleText")
	}
	spots := []map[string]any{}
	if strings.TrimSpace(in.ArticleText) != "" {
		spots = append(spots, map[string]any{"name": "candidate", "context": strings.TrimSpace(in.ArticleText)})
	}
	out, err := common.WriteStageArtifact(strings.TrimSpace(req.OutDir), "extract-spots", identity, "candidates", map[string]any{"identity": identity, "stage": "extractspots", "presenterName": nil, "spots": spots})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "extractspots", OutputPath: out}, nil
}
