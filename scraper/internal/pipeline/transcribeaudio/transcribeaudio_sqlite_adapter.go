package transcribeaudio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/transcription"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(ctx context.Context, req Request) (Response, error) {
	repo, err := repository.New(strings.TrimSpace(req.DBPath))
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}

	src, err := repo.GetLatestArticleAudioSource()
	if err != nil {
		return Response{}, err
	}
	if src == nil {
		return Response{}, fmt.Errorf("no audio source rows with non-empty audio_blob found")
	}

	lang := strings.TrimSpace(req.Language)
	if lang == "" {
		lang = "nl"
	}
	client := transcription.NewMurmelClient(defaultMurmelAPIKey())
	if err := client.Validate(); err != nil {
		return Response{}, err
	}
	filename := fmt.Sprintf("article_audio_source_%d.%s", src.AudioSourceID, cliutil.SafeExt(src.AudioFormat))
	res, err := client.Transcribe(ctx, filename, src.AudioBlob, lang)
	if err != nil {
		return Response{}, err
	}

	msg, jsonErr := canonicalizeJSON(res.Body)
	if jsonErr != nil {
		msg = "{}"
		errMessage := fmt.Sprintf("non-JSON response persisted with fallback payload: %v", jsonErr)
		if res.ErrMessage != nil {
			errMessage = *res.ErrMessage + "; " + errMessage
		}
		res.ErrMessage = &errMessage
	}

	id, err := repo.UpsertArticleAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    src.AudioSourceID,
		Provider:         "murmel",
		Language:         lang,
		HTTPStatus:       res.HTTPStatus,
		ResponseJSON:     msg,
		ResponseByteSize: int64(len(msg)),
		ErrorMessage:     res.ErrMessage,
	})
	if err != nil {
		return Response{}, err
	}

	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = fmt.Sprintf("transcription-%d", id)
	}
	return Response{Identity: identity, Stage: "transcribeaudio", TranscriptionID: id}, nil
}

func defaultMurmelAPIKey() string {
	return strings.TrimSpace(os.Getenv("MURMEL_API_KEY"))
}

func canonicalizeJSON(raw []byte) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return "", fmt.Errorf("empty response body")
	}
	if !json.Valid(raw) {
		return "", fmt.Errorf("response body is not valid JSON")
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return "", fmt.Errorf("canonicalizing JSON: %w", err)
	}
	return buf.String(), nil
}
