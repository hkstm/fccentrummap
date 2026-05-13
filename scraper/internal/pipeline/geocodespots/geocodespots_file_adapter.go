package geocodespots

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type inputPayload struct {
	Query string `json:"query"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(ctx context.Context, req Request) (Response, error) {
	if req.InputPath == "" {
		return Response{}, fmt.Errorf("geocodespots file input requires inputPath")
	}
	identity := common.IdentityFromPath(req.InputPath)
	in, err := common.ReadJSON[inputPayload](req.InputPath)
	if err != nil {
		return Response{}, err
	}
	if strings.TrimSpace(in.Query) == "" {
		return Response{}, fmt.Errorf("geocode input missing required field: query")
	}
	g, err := geocoder.New()
	if err != nil {
		return Response{}, fmt.Errorf("failed to initialize geocoder: %w", err)
	}
	coords, err := g.GeocodePlace(ctx, strings.TrimSpace(in.Query))
	if err != nil {
		return Response{}, err
	}
	out, err := common.WriteStageArtifact("", "geocode-spots", identity, "geocoded", map[string]any{"identity": identity, "stage": "geocodespots", "query": strings.TrimSpace(in.Query), "coordinates": coords})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "geocodespots", OutputPath: out}, nil
}
