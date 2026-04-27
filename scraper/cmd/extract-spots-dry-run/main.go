package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/extraction"
	genaiclient "github.com/hkstm/fccentrummap/internal/genai"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
)

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbPath := flag.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database (SPOTS_DB_PATH overrides default)")
	transcriptionID := flag.Int64("transcription-id", 0, "optional transcription_id to process")
	useLatest := flag.Bool("use-latest", false, "use latest transcription row when --transcription-id is omitted")
	outDir := flag.String("out-dir", cliutil.DefaultDataDir(), "directory for dry-run artifacts")
	gemmaModel := flag.String("gemma-model", defaultGemmaModel(), "Gemma model identifier for generateContent")
	apiKey := flag.String("gemini-api-key", defaultGeminiAPIKey(), "Gemini API key (defaults to GEMINI_API_KEY or GOOGLE_API_KEY)")
	endpoint := flag.String("google-endpoint", strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")), "optional endpoint override for generateContent")
	flag.Parse()

	if *transcriptionID <= 0 && !*useLatest {
		return fmt.Errorf("missing transcription selector: provide --transcription-id <id> or pass --use-latest")
	}

	repo, err := repository.New(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer repo.Close()

	if err := repo.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	row, err := loadTranscriptionRow(repo, *transcriptionID, *useLatest)
	if err != nil {
		return fmt.Errorf("failed to select transcription: %w", err)
	}
	if row == nil {
		if *transcriptionID > 0 {
			return fmt.Errorf("transcription %d not found", *transcriptionID)
		}
		return fmt.Errorf("no transcription rows found")
	}

	sentences, err := extraction.ParseSentenceUnits(row.ResponseJSON)
	if err != nil {
		return fmt.Errorf("transcription %d cannot be used for extraction: %w", row.TranscriptionID, err)
	}

	prompt, err := extraction.BuildDutchPrompt(sentences)
	if err != nil {
		return fmt.Errorf("failed to build Dutch prompt: %w", err)
	}

	client := genaiclient.NewClientWithEndpoint(*apiKey, strings.TrimSpace(*gemmaModel), strings.TrimSpace(*endpoint))
	if err := client.Validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", *outDir, err)
	}

	runTS := time.Now().UTC().Format("20060102T150405Z")
	baseName := fmt.Sprintf("transcript_extraction_%d_%s", row.TranscriptionID, runTS)
	transcriptPath := filepath.Join(*outDir, baseName+"_transcript.json")
	promptPath := filepath.Join(*outDir, baseName+"_prompt.txt")
	responsePath := filepath.Join(*outDir, baseName+"_response.json")
	responseErrorPath := filepath.Join(*outDir, baseName+"_response_error.txt")

	if err := os.WriteFile(transcriptPath, []byte(row.ResponseJSON), 0o644); err != nil {
		return fmt.Errorf("failed to write transcript artifact: %w", err)
	}
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return fmt.Errorf("failed to write prompt artifact: %w", err)
	}

	result, err := client.GenerateContent(context.Background(), prompt, extraction.GenerateContentConfig())
	if err != nil {
		diagnostics := []byte(err.Error() + "\n")
		if writeErr := os.WriteFile(responseErrorPath, diagnostics, 0o644); writeErr != nil {
			return fmt.Errorf("model request failed: %w (also failed to write error artifact %s: %v)", err, responseErrorPath, writeErr)
		}
		return fmt.Errorf("model request failed (error artifact preserved at %s): %w", responseErrorPath, err)
	}

	if err := os.WriteFile(responsePath, result.Body, 0o644); err != nil {
		return fmt.Errorf("failed to write response artifact: %w", err)
	}

	if _, err := extraction.ParseAndValidateResponse(result.Body); err != nil {
		return fmt.Errorf("model response parse/validation failed (raw response preserved at %s): %w", responsePath, err)
	}

	fmt.Printf("transcription_id=%d\n", row.TranscriptionID)
	fmt.Printf("transcript_artifact=%s\n", transcriptPath)
	fmt.Printf("prompt_artifact=%s\n", promptPath)
	fmt.Printf("response_artifact=%s\n", responsePath)
	return nil
}

func loadTranscriptionRow(repo *repository.Repository, transcriptionID int64, useLatest bool) (*models.ArticleAudioTranscription, error) {
	if transcriptionID > 0 {
		return repo.GetArticleAudioTranscriptionByID(transcriptionID)
	}
	if !useLatest {
		return nil, nil
	}
	return repo.GetLatestArticleAudioTranscription()
}

func defaultGemmaModel() string {
	if model := strings.TrimSpace(os.Getenv("GEMMA_MODEL")); model != "" {
		return model
	}
	return "gemma-4-31b-it"
}

func defaultGeminiAPIKey() string {
	if v := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_API_KEY")); v != "" {
		return v
	}
	return ""
}
