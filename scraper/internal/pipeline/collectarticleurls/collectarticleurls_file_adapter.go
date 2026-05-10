package collectarticleurls

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type inputPayload struct {
	ArticleURL string   `json:"articleUrl"`
	URLs       []string `json:"urls"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, req Request) (Response, error) {
	payload := inputPayload{ArticleURL: strings.TrimSpace(req.ArticleURL)}
	if strings.TrimSpace(req.InputPath) != "" {
		in, err := common.ReadJSON[inputPayload](req.InputPath)
		if err != nil {
			return Response{}, err
		}
		payload = in
		if u := strings.TrimSpace(req.ArticleURL); u != "" {
			if strings.TrimSpace(payload.ArticleURL) == "" {
				payload.ArticleURL = u
			} else if payload.ArticleURL != u {
				payload.URLs = append(payload.URLs, u)
			}
		}
	}

	identity := strings.TrimSpace(req.Identity)
	if identity == "" && strings.TrimSpace(req.InputPath) != "" {
		identity = common.IdentityFromPath(req.InputPath)
	}
	if identity == "" {
		identity = common.IdentityFromURL(payload.ArticleURL)
	}
	if identity == "" {
		return Response{}, fmt.Errorf("collectarticleurls file input requires identity when inputPath is omitted")
	}

	seen := map[string]struct{}{}
	urls := make([]string, 0, len(payload.URLs)+1)
	if u := strings.TrimSpace(payload.ArticleURL); u != "" {
		if _, ok := seen[u]; !ok {
			seen[u] = struct{}{}
			urls = append(urls, u)
		}
	}
	for _, u := range payload.URLs {
		trimmed := strings.TrimSpace(u)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		urls = append(urls, trimmed)
	}
	if len(urls) == 0 {
		return Response{}, fmt.Errorf("collectarticleurls file input requires articleUrl or urls")
	}
	out, err := common.WriteStageArtifact("", "collect-article-urls", identity, "articles", map[string]any{"identity": identity, "stage": "collectarticleurls", "articleUrls": urls})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "collectarticleurls", URLs: urls, OutputPath: out}, nil
}
