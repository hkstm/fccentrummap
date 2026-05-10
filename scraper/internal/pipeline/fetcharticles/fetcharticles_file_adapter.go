package fetcharticles

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type inputPayload struct {
	ArticleURLs []string `json:"articleUrls"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, req Request) (Response, error) {
	if strings.TrimSpace(req.InputPath) == "" {
		return Response{}, fmt.Errorf("fetcharticles file input requires inputPath")
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = common.IdentityFromPath(req.InputPath)
	}
	in, err := common.ReadJSON[inputPayload](req.InputPath)
	if err != nil {
		return Response{}, err
	}
	if len(in.ArticleURLs) == 0 {
		return Response{}, fmt.Errorf("fetcharticles input missing articleUrls")
	}
	out, err := common.WriteStageArtifact("", "fetch-articles", identity, "fetched", map[string]any{"identity": identity, "stage": "fetcharticles", "articleUrls": in.ArticleURLs, "fetchedCount": len(in.ArticleURLs)})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "fetcharticles", ArticleURLs: in.ArticleURLs, FetchedCount: len(in.ArticleURLs), OutputPath: out}, nil
}
