package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/transcription"
)

func main() {
	log.SetFlags(0)

	dbPath := flag.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database (SPOTS_DB_PATH overrides default)")
	audioSourceID := flag.Int64("audio-source-id", 0, "optional audio_source_id to transcribe; latest row is used when omitted")
	language := flag.String("language", "nl", "language code sent to Murmel")
	flag.Parse()

	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	if err := repo.InitSchema(); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	var src *models.ArticleAudioSource
	if *audioSourceID > 0 {
		src, err = repo.GetArticleAudioSourceByID(*audioSourceID)
		if err != nil {
			log.Fatalf("failed to load audio source %d: %v", *audioSourceID, err)
		}
		if src == nil {
			log.Fatalf("audio source %d not found", *audioSourceID)
		}
	} else {
		src, err = repo.GetLatestArticleAudioSource()
		if err != nil {
			log.Fatalf("failed to load latest audio source: %v", err)
		}
		if src == nil {
			log.Fatalf("no audio source rows with non-empty audio_blob found")
		}
	}

	apiClient := transcription.NewMurmelClient(os.Getenv("MURMEL_API_KEY"))
	if err := apiClient.Validate(); err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("article_audio_source_%d.%s", src.AudioSourceID, cliutil.SafeExt(src.AudioFormat))
	result, err := apiClient.Transcribe(context.Background(), filename, src.AudioBlob, *language)
	if err != nil {
		log.Fatalf("transcription request failed: %v", err)
	}

	canonicalJSON, err := canonicalizeJSON(result.Body)
	if err != nil {
		canonicalJSON = "{}"
		msg := fmt.Sprintf("non-JSON response persisted with fallback payload: %v", err)
		if result.ErrMessage != nil {
			msg = *result.ErrMessage + "; " + msg
		}
		result.ErrMessage = &msg
	}

	row := models.ArticleAudioTranscription{
		AudioSourceID:    src.AudioSourceID,
		Provider:         "murmel",
		Language:         *language,
		HTTPStatus:       result.HTTPStatus,
		ResponseJSON:     canonicalJSON,
		ResponseByteSize: int64(len(canonicalJSON)),
		ErrorMessage:     result.ErrMessage,
	}

	transcriptionID, err := repo.UpsertArticleAudioTranscription(row)
	if err != nil {
		log.Fatalf("failed to persist transcription result: %v", err)
	}

	fmt.Printf("audio_source_id=%d\n", src.AudioSourceID)
	fmt.Printf("murmel_http_status=%d\n", result.HTTPStatus)
	if result.ErrMessage != nil {
		fmt.Printf("murmel_error=%s\n", *result.ErrMessage)
	}
	fmt.Printf("transcription_id=%d\n", transcriptionID)
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
