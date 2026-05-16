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
	repo, err := repository.New(req.DBPath)
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}

	rows, err := repo.ListAudioSourcesPendingTranscription()
	if err != nil {
		return Response{}, err
	}
	if len(rows) == 0 {
		return Response{Identity: "transcriptions-none", Stage: "transcribeaudio"}, nil
	}

	var processedIDs []int64
	for i, row := range rows {
		fmt.Printf("Processing row %d/%d (AudioSourceID=%d)\n", i+1, len(rows), row.AudioSourceID)
		id, err := processRow(repo, row, req, ctx)
		if err != nil {
			return Response{}, fmt.Errorf("processing audio source %d: %w", row.AudioSourceID, err)
		}
		processedIDs = append(processedIDs, id)
	}

	return Response{
		Identity:         fmt.Sprintf("transcriptions-batch-%d", processedIDs[len(processedIDs)-1]),
		Stage:            "transcribeaudio",
		TranscriptionIDs: processedIDs,
	}, nil
}

func processRow(repo *repository.Repository, src models.ArticleAudioSource, req Request, ctx context.Context) (int64, error) {
	lang := req.Language
	if lang == "" {
		lang = "nl"
	}
	client := transcription.NewMurmelClient(defaultMurmelAPIKey())
	if err := client.Validate(); err != nil {
		return 0, err
	}
	filename := fmt.Sprintf("article_audio_source_%d.%s", src.AudioSourceID, cliutil.SafeExt(src.AudioFormat))
	res, err := client.Transcribe(ctx, filename, src.AudioBlob, lang)
	if err != nil {
		return 0, err
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

	id, err := repo.UpsertAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    src.AudioSourceID,
		Provider:         "murmel",
		Language:         lang,
		HTTPStatus:       res.HTTPStatus,
		ResponseJSON:     msg,
		ResponseByteSize: int64(len(msg)),
		ErrorMessage:     res.ErrMessage,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
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
