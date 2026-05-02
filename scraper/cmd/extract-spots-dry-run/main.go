package main

import (
	"context"
	"encoding/json"
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
	scraperpkg "github.com/hkstm/fccentrummap/internal/scraper"
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
	articleURL := flag.String("article-url", "", "optional article URL; when set selects latest transcription linked to this article")
	useLatest := flag.Bool("use-latest", false, "use latest transcription row when --transcription-id and --article-url are omitted")
	outDir := flag.String("out-dir", cliutil.DefaultDataDir(), "directory for dry-run artifacts")
	gemmaModel := flag.String("gemma-model", defaultGemmaModel(), "Gemma model identifier for generateContent")
	apiKey := flag.String("gemini-api-key", defaultGeminiAPIKey(), "Gemini API key (defaults to GEMINI_API_KEY or GOOGLE_API_KEY)")
	endpoint := flag.String("google-endpoint", strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")), "optional endpoint override for generateContent")
	resetExtractionStorage := flag.Bool("reset-extraction-storage", false, "drop/recreate spot extraction storage table before running and keep backup table")
	flag.Parse()

	if *transcriptionID <= 0 && strings.TrimSpace(*articleURL) == "" && !*useLatest {
		return fmt.Errorf("missing selector: provide --transcription-id <id>, --article-url <url>, or pass --use-latest")
	}
	if *transcriptionID > 0 && strings.TrimSpace(*articleURL) != "" {
		return fmt.Errorf("choose either --transcription-id or --article-url, not both")
	}

	repo, err := repository.New(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer repo.Close()

	if err := repo.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	if *resetExtractionStorage {
		backupTable, err := repo.ResetSpotExtractionStorageWithBackup("")
		if err != nil {
			return fmt.Errorf("failed to reset extraction storage: %w", err)
		}
		if backupTable == "" {
			log.Printf("INFO: spot extraction storage reset (no prior table found)")
		} else {
			log.Printf("INFO: spot extraction storage reset; backup preserved in table %s", backupTable)
		}
	}

	row, err := loadTranscriptionRow(repo, *transcriptionID, strings.TrimSpace(*articleURL), *useLatest)
	if err != nil {
		return fmt.Errorf("failed to select transcription: %w", err)
	}
	if row == nil {
		if *transcriptionID > 0 {
			return fmt.Errorf("transcription %d not found", *transcriptionID)
		}
		if strings.TrimSpace(*articleURL) != "" {
			return fmt.Errorf("no transcription rows found for article URL %s", strings.TrimSpace(*articleURL))
		}
		return fmt.Errorf("no transcription rows found")
	}

	audioSource, err := repo.GetArticleAudioSourceByID(row.AudioSourceID)
	if err != nil {
		return fmt.Errorf("failed to load audio source for transcription %d: %w", row.TranscriptionID, err)
	}
	if audioSource == nil {
		return fmt.Errorf("audio source %d linked from transcription %d not found", row.AudioSourceID, row.TranscriptionID)
	}

	articleRawID := audioSource.ArticleRawID
	var articleRaw *models.ArticleRaw
	if url := strings.TrimSpace(*articleURL); url != "" {
		articleRaw, err = repo.GetArticleRawByURL(url)
		if err != nil {
			return fmt.Errorf("failed to load article raw by URL: %w", err)
		}
	} else {
		articleRaw, err = repo.GetArticleRawByID(articleRawID)
		if err != nil {
			return fmt.Errorf("failed to load article_raw_id=%d: %w", articleRawID, err)
		}
	}
	if articleRaw == nil {
		return fmt.Errorf("article_raw_id=%d not found", articleRawID)
	}

	cleanedArticleText, err := loadCleanedArticleText(repo, articleRaw)
	if err != nil {
		return fmt.Errorf("failed to load cleaned article text article_raw_id=%d: %w", articleRaw.ArticleRawID, err)
	}

	sentences, err := extraction.ParseSentenceUnits(row.ResponseJSON)
	if err != nil {
		return fmt.Errorf("transcription %d cannot be used for extraction: %w", row.TranscriptionID, err)
	}

	pass1Prompt, err := extraction.BuildDutchPass1Prompt(extraction.PromptInput{
		CleanedArticleText: cleanedArticleText,
		Sentences:          sentences,
	})
	if err != nil {
		return fmt.Errorf("failed to build Dutch pass-1 prompt: %w", err)
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
	articlePath := filepath.Join(*outDir, baseName+"_article.txt")
	pass1PromptPath := filepath.Join(*outDir, baseName+"_pass1_prompt.txt")
	pass2PromptPath := filepath.Join(*outDir, baseName+"_pass2_prompt.txt")
	pass1ResponsePath := filepath.Join(*outDir, baseName+"_pass1_response.json")
	pass2ResponsePath := filepath.Join(*outDir, baseName+"_pass2_response.json")
	pass1ErrorPath := filepath.Join(*outDir, baseName+"_pass1_response_error.txt")
	pass2ErrorPath := filepath.Join(*outDir, baseName+"_pass2_response_error.txt")

	if err := os.WriteFile(transcriptPath, []byte(row.ResponseJSON), 0o644); err != nil {
		return fmt.Errorf("failed to write transcript artifact: %w", err)
	}
	if err := os.WriteFile(articlePath, []byte(cleanedArticleText), 0o644); err != nil {
		return fmt.Errorf("failed to write cleaned article artifact: %w", err)
	}
	if err := os.WriteFile(pass1PromptPath, []byte(pass1Prompt), 0o644); err != nil {
		return fmt.Errorf("failed to write pass-1 prompt artifact: %w", err)
	}

	pass1Result, err := client.GenerateContent(context.Background(), pass1Prompt, extraction.GeneratePass1ContentConfig())
	if err != nil {
		diagnostics := []byte(err.Error() + "\n")
		if writeErr := os.WriteFile(pass1ErrorPath, diagnostics, 0o644); writeErr != nil {
			return fmt.Errorf("pass-1 model request failed: %w (also failed to write error artifact %s: %v)", err, pass1ErrorPath, writeErr)
		}
		return fmt.Errorf("pass-1 model request failed (error artifact preserved at %s): %w", pass1ErrorPath, err)
	}
	if err := os.WriteFile(pass1ResponsePath, pass1Result.Body, 0o644); err != nil {
		return fmt.Errorf("failed to write pass-1 response artifact: %w", err)
	}

	pass1Parsed, err := extraction.ParseAndValidateResponse(pass1Result.Body)
	if err != nil {
		return fmt.Errorf("pass-1 model response parse/validation failed (raw response preserved at %s): %w", pass1ResponsePath, err)
	}

	var (
		pass2PromptText string
		pass2RawBody    []byte
		pass2Parsed     *extraction.ParsedRefinementResponse
	)

	if len(pass1Parsed.Spots) > 0 {
		pass2PromptText, err = extraction.BuildDutchPass2RefinementPrompt(extraction.RefinementPromptInput{
			Sentences:  sentences,
			Pass1Spots: pass1Parsed.Spots,
		})
		if err != nil {
			return fmt.Errorf("failed to build Dutch pass-2 prompt: %w", err)
		}
		if err := os.WriteFile(pass2PromptPath, []byte(pass2PromptText), 0o644); err != nil {
			return fmt.Errorf("failed to write pass-2 prompt artifact: %w", err)
		}

		pass2Result, err := client.GenerateContent(context.Background(), pass2PromptText, extraction.GeneratePass2RefinementContentConfig())
		if err != nil {
			diagnostics := []byte(err.Error() + "\n")
			if writeErr := os.WriteFile(pass2ErrorPath, diagnostics, 0o644); writeErr != nil {
				return fmt.Errorf("pass-2 model request failed: %w (also failed to write error artifact %s: %v)", err, pass2ErrorPath, writeErr)
			}
			return fmt.Errorf("pass-2 model request failed (error artifact preserved at %s): %w", pass2ErrorPath, err)
		}
		pass2RawBody = pass2Result.Body
		pass2Parsed, err = extraction.ParseAndValidateRefinementResponse(pass2RawBody, pass1Parsed.Spots)
		if err != nil {
			return fmt.Errorf("pass-2 model response parse/validation failed: %w", err)
		}
		if err := os.WriteFile(pass2ResponsePath, pass2RawBody, 0o644); err != nil {
			return fmt.Errorf("failed to write pass-2 response artifact: %w", err)
		}
	} else {
		pass2PromptText = "SKIPPED: no pass-1 spots available for refinement\n"
		if err := os.WriteFile(pass2PromptPath, []byte(pass2PromptText), 0o644); err != nil {
			return fmt.Errorf("failed to write skipped pass-2 prompt artifact: %w", err)
		}
		pass2RawBody = []byte(`{"skipped":true,"reason":"no pass-1 spots"}`)
		if err := os.WriteFile(pass2ResponsePath, pass2RawBody, 0o644); err != nil {
			return fmt.Errorf("failed to write skipped pass-2 response artifact: %w", err)
		}
	}

	finalParsed := extraction.ApplyRefinements(pass1Parsed, pass2Parsed)
	if _, err := json.Marshal(finalParsed); err != nil {
		return fmt.Errorf("failed to marshal final parsed response: %w", err)
	}

	fmt.Printf("article_url=%s\n", articleRaw.URL)
	fmt.Printf("transcription_id=%d\n", row.TranscriptionID)
	fmt.Printf("presenter_name=%s\n", nullableString(finalParsed.PresenterName))
	fmt.Printf("article_artifact=%s\n", articlePath)
	fmt.Printf("transcript_artifact=%s\n", transcriptPath)
	fmt.Printf("pass1_prompt_artifact=%s\n", pass1PromptPath)
	fmt.Printf("pass2_prompt_artifact=%s\n", pass2PromptPath)
	fmt.Printf("pass1_response_artifact=%s\n", pass1ResponsePath)
	fmt.Printf("pass2_response_artifact=%s\n", pass2ResponsePath)
	return nil
}

func loadTranscriptionRow(repo *repository.Repository, transcriptionID int64, articleURL string, useLatest bool) (*models.ArticleAudioTranscription, error) {
	if transcriptionID > 0 {
		return repo.GetArticleAudioTranscriptionByID(transcriptionID)
	}
	if articleURL != "" {
		return repo.GetLatestArticleAudioTranscriptionByURL(articleURL)
	}
	if !useLatest {
		return nil, nil
	}
	return repo.GetLatestArticleAudioTranscription()
}

func loadCleanedArticleText(repo *repository.Repository, articleRaw *models.ArticleRaw) (string, error) {
	contents, err := repo.ListArticleTextContents(articleRaw.ArticleRawID)
	if err != nil {
		return "", err
	}
	if text := joinPersistedTextContents(contents); text != "" {
		return text, nil
	}

	fallback := scraperpkg.ExtractArticleTextContent(articleRaw.HTML)
	if text := joinExtractedTextContents(fallback.Contents); text != "" {
		return text, nil
	}
	return "", fmt.Errorf("article text extraction has no non-empty cleaned content")
}

func joinPersistedTextContents(contents []models.ArticleTextContent) string {
	parts := make([]string, 0, len(contents))
	for _, content := range contents {
		if trimmed := strings.TrimSpace(content.Content); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "\n\n")
}

func joinExtractedTextContents(contents []models.ArticleTextContentInput) string {
	parts := make([]string, 0, len(contents))
	for _, content := range contents {
		if trimmed := strings.TrimSpace(content.Content); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "\n\n")
}

func nullableString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
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
