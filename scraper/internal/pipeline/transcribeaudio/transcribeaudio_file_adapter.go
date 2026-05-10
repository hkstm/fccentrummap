package transcribeaudio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/pipeline/common"
)

type FileAdapter struct{}

type inputPayload struct {
	AudioSourceID int64  `json:"audioSourceId"`
	Language      string `json:"language"`
	AudioFormat   string `json:"audioFormat"`
	AudioBlobB64  string `json:"audioBlobBase64"`
}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, req Request) (Response, error) {
	if strings.TrimSpace(req.InputPath) == "" {
		return Response{}, fmt.Errorf("transcribeaudio file input requires inputPath")
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = common.IdentityFromPath(req.InputPath)
	}
	in, err := common.ReadJSON[inputPayload](req.InputPath)
	if err != nil {
		return Response{}, err
	}
	if in.AudioSourceID <= 0 {
		return Response{}, fmt.Errorf("transcribeaudio input missing audioSourceId")
	}
	if strings.TrimSpace(in.AudioBlobB64) == "" {
		return Response{}, fmt.Errorf("transcribeaudio input missing audioBlobBase64")
	}
	lang := strings.TrimSpace(in.Language)
	if lang == "" {
		lang = "nl"
	}
	out, err := common.WriteStageArtifact("", "transcribe-audio", identity, "transcription", map[string]any{"identity": identity, "stage": "transcribeaudio", "provider": "murmel", "language": lang, "httpStatus": 0, "responseJson": "{}"})
	if err != nil {
		return Response{}, err
	}
	return Response{Identity: identity, Stage: "transcribeaudio", OutputPath: out}, nil
}
