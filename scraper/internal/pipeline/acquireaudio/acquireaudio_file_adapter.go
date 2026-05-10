package acquireaudio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type article struct {
	URL     string `json:"url"`
	VideoID string `json:"videoId"`
}

type inputPayload struct {
	Articles []article `json:"articles"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, req Request) (Response, error) {
	if strings.TrimSpace(req.InputPath) == "" {
		return Response{}, fmt.Errorf("acquireaudio file input requires inputPath")
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = common.IdentityFromPath(req.InputPath)
	}
	in, err := common.ReadJSON[inputPayload](req.InputPath)
	if err != nil {
		return Response{}, err
	}
	if len(in.Articles) == 0 {
		return Response{}, fmt.Errorf("acquireaudio input missing articles")
	}
	acquired := make([]AcquiredAudio, 0, len(in.Articles))
	for _, a := range in.Articles {
		if strings.TrimSpace(a.URL) == "" || strings.TrimSpace(a.VideoID) == "" {
			return Response{}, fmt.Errorf("acquireaudio input contains article without url/videoId")
		}
		acquired = append(acquired, AcquiredAudio{URL: strings.TrimSpace(a.URL), VideoID: strings.TrimSpace(a.VideoID)})
	}
	out, err := common.WriteStageArtifact("", "acquire-audio", identity, "audio", map[string]any{"identity": identity, "stage": "acquireaudio", "acquired": acquired})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "acquireaudio", Acquired: acquired, OutputPath: out}, nil
}
