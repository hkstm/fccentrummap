package extractspots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hkstm/fccentrummap/internal/articletext"
	"github.com/hkstm/fccentrummap/internal/extraction"
	genaiclient "github.com/hkstm/fccentrummap/internal/genai"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type Options struct {
	TranscriptionID int64
	ArticleURL      string
	UseLatest       bool
	OutDir          string
	GemmaModel      string
	APIKey          string
	Endpoint        string
	PersistRecord   bool
}

type Result struct {
	ArticleURL             string
	TranscriptionID        int64
	PresenterName          *string
	ArticleArtifactPath    string
	TranscriptArtifactPath string
	Pass1PromptPath        string
	Pass2PromptPath        string
	Pass1ResponsePath      string
	Pass2ResponsePath      string
	SpotExtractionID       int64
}

func Run(ctx context.Context, repo *repository.Repository, opts Options) (*Result, error) {
	articleURL := strings.TrimSpace(opts.ArticleURL)
	if opts.TranscriptionID <= 0 && articleURL == "" && !opts.UseLatest {
		return nil, fmt.Errorf("missing selector: provide transcription id, article url, or use latest")
	}
	if opts.TranscriptionID > 0 && articleURL != "" {
		return nil, fmt.Errorf("choose either transcription id or article url, not both")
	}

	row, err := loadTranscriptionRow(repo, opts.TranscriptionID, articleURL, opts.UseLatest)
	if err != nil {
		return nil, fmt.Errorf("failed to select transcription: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("no transcription rows found")
	}

	audioSource, err := repo.GetArticleAudioSourceByID(row.AudioSourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio source for transcription %d: %w", row.TranscriptionID, err)
	}
	if audioSource == nil {
		return nil, fmt.Errorf("audio source %d linked from transcription %d not found", row.AudioSourceID, row.TranscriptionID)
	}

	var articleRaw *models.ArticleRaw
	if articleURL != "" {
		articleRaw, err = repo.GetArticleRawByURL(articleURL)
	} else {
		articleRaw, err = repo.GetArticleRawByID(audioSource.ArticleRawID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load article row: %w", err)
	}
	if articleRaw == nil {
		return nil, fmt.Errorf("article_raw_id=%d not found", audioSource.ArticleRawID)
	}

	cleanedArticleText, err := loadCleanedArticleText(repo, articleRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to load cleaned article text article_raw_id=%d: %w", articleRaw.ArticleRawID, err)
	}

	sentences, err := extraction.ParseSentenceUnits(row.ResponseJSON)
	if err != nil {
		return nil, fmt.Errorf("transcription %d cannot be used for extraction: %w", row.TranscriptionID, err)
	}

	pass1Prompt, err := extraction.BuildDutchPass1Prompt(extraction.PromptInput{CleanedArticleText: cleanedArticleText, Sentences: sentences})
	if err != nil {
		return nil, fmt.Errorf("failed to build pass-1 prompt: %w", err)
	}

	client := genaiclient.NewClientWithEndpoint(strings.TrimSpace(opts.APIKey), strings.TrimSpace(opts.GemmaModel), strings.TrimSpace(opts.Endpoint))
	if err := client.Validate(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", opts.OutDir, err)
	}

	runTS := time.Now().UTC().Format("20060102T150405Z")
	baseName := fmt.Sprintf("transcript_extraction_%d_%s", row.TranscriptionID, runTS)
	transcriptPath := filepath.Join(opts.OutDir, baseName+"_transcript.json")
	articlePath := filepath.Join(opts.OutDir, baseName+"_article.txt")
	pass1PromptPath := filepath.Join(opts.OutDir, baseName+"_pass1_prompt.txt")
	pass2PromptPath := filepath.Join(opts.OutDir, baseName+"_pass2_prompt.txt")
	pass1ResponsePath := filepath.Join(opts.OutDir, baseName+"_pass1_response.json")
	pass2ResponsePath := filepath.Join(opts.OutDir, baseName+"_pass2_response.json")

	if err := os.WriteFile(transcriptPath, []byte(row.ResponseJSON), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(articlePath, []byte(cleanedArticleText), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(pass1PromptPath, []byte(pass1Prompt), 0o644); err != nil {
		return nil, err
	}

	pass1Result, err := client.GenerateContent(ctx, pass1Prompt, extraction.GeneratePass1ContentConfig())
	if err != nil {
		return nil, fmt.Errorf("pass-1 model request failed: %w", err)
	}
	if err := os.WriteFile(pass1ResponsePath, pass1Result.Body, 0o644); err != nil {
		return nil, err
	}
	pass1Parsed, err := extraction.ParseAndValidateResponse(pass1Result.Body)
	if err != nil {
		return nil, fmt.Errorf("pass-1 parse/validation failed: %w", err)
	}

	var pass2Parsed *extraction.ParsedRefinementResponse
	pass2RawBody := []byte(`{"skipped":true,"reason":"no pass-1 spots"}`)
	if len(pass1Parsed.Spots) > 0 {
		pass2PromptText, err := extraction.BuildDutchPass2RefinementPrompt(extraction.RefinementPromptInput{Sentences: sentences, Pass1Spots: pass1Parsed.Spots})
		if err != nil {
			return nil, fmt.Errorf("failed to build pass-2 prompt: %w", err)
		}
		if err := os.WriteFile(pass2PromptPath, []byte(pass2PromptText), 0o644); err != nil {
			return nil, err
		}
		pass2Result, err := client.GenerateContent(ctx, pass2PromptText, extraction.GeneratePass2RefinementContentConfig())
		if err != nil {
			return nil, fmt.Errorf("pass-2 model request failed: %w", err)
		}
		pass2RawBody = pass2Result.Body
		if err := os.WriteFile(pass2ResponsePath, pass2RawBody, 0o644); err != nil {
			return nil, err
		}
		pass2Parsed, err = extraction.ParseAndValidateRefinementResponse(pass2RawBody, pass1Parsed.Spots)
		if err != nil {
			return nil, fmt.Errorf("pass-2 parse/validation failed (raw response preserved at %s): %w", pass2ResponsePath, err)
		}
	} else {
		if err := os.WriteFile(pass2PromptPath, []byte("SKIPPED: no pass-1 spots available for refinement\n"), 0o644); err != nil {
			return nil, err
		}
		if err := os.WriteFile(pass2ResponsePath, pass2RawBody, 0o644); err != nil {
			return nil, err
		}
	}

	finalParsed := extraction.ApplyRefinements(pass1Parsed, pass2Parsed)
	parsedBytes, err := json.Marshal(finalParsed)
	if err != nil {
		return nil, fmt.Errorf("marshal final parsed response: %w", err)
	}
	var spotExtractionID int64
	if opts.PersistRecord {
		spotExtractionID, err = repo.InsertSpotExtractionRecord(models.SpotExtractionRecordInput{
			ArticleRawID:       articleRaw.ArticleRawID,
			TranscriptionID:    row.TranscriptionID,
			PresenterName:      finalParsed.PresenterName,
			PromptText:         pass1Prompt,
			RawResponseJSON:    string(pass1Result.Body),
			ParsedResponseJSON: string(parsedBytes),
		})
		if err != nil {
			return nil, fmt.Errorf("persist extraction record: %w", err)
		}
	}

	return &Result{
		ArticleURL:             articleRaw.URL,
		TranscriptionID:        row.TranscriptionID,
		PresenterName:          finalParsed.PresenterName,
		ArticleArtifactPath:    articlePath,
		TranscriptArtifactPath: transcriptPath,
		Pass1PromptPath:        pass1PromptPath,
		Pass2PromptPath:        pass2PromptPath,
		Pass1ResponsePath:      pass1ResponsePath,
		Pass2ResponsePath:      pass2ResponsePath,
		SpotExtractionID:       spotExtractionID,
	}, nil
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
	parts := make([]string, 0, len(contents))
	for _, content := range contents {
		if trimmed := strings.TrimSpace(content.Content); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, "\n\n"), nil
	}

	fallback := articletext.ExtractArticleTextContent(articleRaw.HTML)
	parts = make([]string, 0, len(fallback.Contents))
	for _, c := range fallback.Contents {
		if trimmed := strings.TrimSpace(c.Content); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, "\n\n"), nil
	}
	return "", fmt.Errorf("article text extraction has no non-empty cleaned content")
}
