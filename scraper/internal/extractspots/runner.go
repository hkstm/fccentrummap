package extractspots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hkstm/fccentrummap/internal/extraction"
	genaiclient "github.com/hkstm/fccentrummap/internal/genai"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type Options struct {
	OutDir     string
	GemmaModel string
	APIKey     string
	Endpoint   string
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
	row, err := repo.GetLatestAudioTranscription()
	if err != nil {
		return nil, fmt.Errorf("failed to select transcription: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("no transcription rows found")
	}

	articleURL, articleSourceID, cleanedArticleText, err := repo.GetArticleContextByTranscriptionID(row.TranscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load article context for transcription %d: %w", row.TranscriptionID, err)
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
	if _, err := json.Marshal(finalParsed); err != nil {
		return nil, fmt.Errorf("marshal final parsed response: %w", err)
	}
	for _, spot := range finalParsed.Spots {
		spotMentionID, err := repo.UpsertSpotMention(row.TranscriptionID, spot.Place, spot.SentenceStartTimestamp, spot.OriginalSentenceStartTimestamp, spot.RefinedSentenceStartTimestamp)
		if err != nil {
			return nil, fmt.Errorf("persist spot mention place=%s: %w", spot.Place, err)
		}
		_ = spotMentionID
	}
	if finalParsed.PresenterName != nil {
		presenterID, err := repo.UpsertPresenter(*finalParsed.PresenterName)
		if err != nil {
			return nil, fmt.Errorf("persist presenter %s: %w", *finalParsed.PresenterName, err)
		}
		if err := repo.LinkArticlePresenter(articleSourceID, presenterID); err != nil {
			return nil, fmt.Errorf("persist article presenter link: %w", err)
		}
	}

	return &Result{
		ArticleURL:             articleURL,
		TranscriptionID:        row.TranscriptionID,
		PresenterName:          finalParsed.PresenterName,
		ArticleArtifactPath:    articlePath,
		TranscriptArtifactPath: transcriptPath,
		Pass1PromptPath:        pass1PromptPath,
		Pass2PromptPath:        pass2PromptPath,
		Pass1ResponsePath:      pass1ResponsePath,
		Pass2ResponsePath:      pass2ResponsePath,
		SpotExtractionID:       row.TranscriptionID,
	}, nil
}

