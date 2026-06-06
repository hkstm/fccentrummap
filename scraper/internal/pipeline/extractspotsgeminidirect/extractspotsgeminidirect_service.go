package extractspotsgeminidirect

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hkstm/fccentrummap/internal/geminidirect"
	gogenai "google.golang.org/genai"
)

type Service struct {
	sqlite SQLitePort
	file   FilePort
}

func NewService(sqlite SQLitePort, file FilePort) *Service {
	return &Service{sqlite: sqlite, file: file}
}

func (s *Service) Run(ctx context.Context, mode string, req Request) (Response, error) {
	switch mode {
	case "sqlite":
		if s.sqlite == nil {
			return Response{}, fmt.Errorf("sqlite adapter not configured")
		}
		return s.sqlite.Run(ctx, req)
	case "file":
		if s.file == nil {
			return Response{}, fmt.Errorf("file adapter not configured")
		}
		return s.file.Run(ctx, req)
	default:
		return Response{}, fmt.Errorf("unsupported mode: %s", mode)
	}
}

type Runner struct {
	Inputs    InputLoader
	Client    ModelClient
	Writer    geminidirect.ArtifactWriter
	Persister AcceptedSpotPersister
	Model     string
	Force     bool
	Limit     int
}

func (r Runner) Run(ctx context.Context) (Response, error) {
	if r.Inputs == nil {
		return Response{}, fmt.Errorf("input loader is required")
	}
	if r.Client == nil {
		return Response{}, fmt.Errorf("Gemini client is required")
	}
	inputs, err := r.Inputs.ListGeminiDirectInputs()
	if err != nil {
		return Response{}, err
	}

	res := Response{Identity: "gemini-direct-none", Stage: "extractspotsgeminidirect", OutputDir: r.Writer.OutDir}
	attempted := 0
	for _, input := range inputs {
		paths := geminidirect.PathsForArticle(r.Writer.OutDir, input.ArticleSourceID)
		articleResult := ArticleResult{
			ArticleSourceID: input.ArticleSourceID,
			ArticleURL:      input.ArticleURL,
			YouTubeURL:      input.YouTubeURL,
			PromptPath:      paths.Prompt,
			RawResponsePath: paths.Raw,
			ParsedPath:      paths.Parsed,
		}
		missing := missingInputs(input)
		if len(missing) > 0 {
			articleResult.Skipped = true
			articleResult.Diagnostics = append(articleResult.Diagnostics, fmt.Sprintf("missing required input(s): %s", strings.Join(missing, ", ")))
			res.SkippedCount++
			res.Articles = append(res.Articles, articleResult)
			continue
		}

		if !r.Force && fileExists(paths.Parsed) {
			articleResult.Cached = true
			articleResult.Skipped = true
			articleResult.Diagnostics = append(articleResult.Diagnostics, "existing parsed artifact found; skipping Gemini request (use --force to regenerate)")
			res.CachedCount++
			res.SkippedCount++
			res.Articles = append(res.Articles, articleResult)
			continue
		}

		if r.Limit > 0 && attempted >= r.Limit {
			break
		}

		prompt, err := geminidirect.BuildPrompt(geminidirect.PromptInput{ArticleURL: input.ArticleURL, YouTubeURL: input.YouTubeURL})
		if err != nil {
			articleResult.Skipped = true
			articleResult.Diagnostics = append(articleResult.Diagnostics, err.Error())
			res.SkippedCount++
			res.Articles = append(res.Articles, articleResult)
			continue
		}
		paths, err = r.Writer.WritePrompt(input.ArticleSourceID, prompt)
		if err != nil {
			return res, err
		}

		attempted++
		parts := []*gogenai.Part{
			gogenai.NewPartFromURI(input.YouTubeURL, "video/*"),
			gogenai.NewPartFromText(prompt),
		}
		modelResult, err := r.Client.GenerateContentWithParts(ctx, parts, geminidirect.GenerateContentConfig())
		if err != nil {
			articleResult.Diagnostics = append(articleResult.Diagnostics, fmt.Sprintf("Gemini request failed: %v", err))
			artifact := &geminidirect.ParsedArtifact{
				ArticleSourceID: input.ArticleSourceID,
				ArticleURL:      strings.TrimSpace(input.ArticleURL),
				YouTubeURL:      strings.TrimSpace(input.YouTubeURL),
				Model:           strings.TrimSpace(r.Model),
				Diagnostics:     articleResult.Diagnostics,
			}
			if _, writeErr := r.Writer.WriteParsed(input.ArticleSourceID, artifact); writeErr != nil {
				return res, writeErr
			}
			res.Articles = append(res.Articles, articleResult)
			continue
		}
		if _, err := r.Writer.WriteRaw(input.ArticleSourceID, modelResult.Body); err != nil {
			return res, err
		}

		artifact, parseErr := geminidirect.ParseRawResponseToArtifact(modelResult.Body, input.ArticleSourceID, input.ArticleURL, input.YouTubeURL, r.Model)
		if parseErr != nil {
			articleResult.Diagnostics = append(articleResult.Diagnostics, parseErr.Error())
		}
		if _, err := r.Writer.WriteParsed(input.ArticleSourceID, artifact); err != nil {
			return res, err
		}
		if parseErr == nil && r.Persister != nil {
			if err := r.Persister.PersistAcceptedGeminiDirectArticle(acceptedArticleFromArtifact(artifact)); err != nil {
				articleResult.Diagnostics = append(articleResult.Diagnostics, fmt.Sprintf("persist accepted Gemini-direct spots: %v", err))
				res.Articles = append(res.Articles, articleResult)
				return res, err
			}
		}
		res.ProcessedCount++
		res.Identity = fmt.Sprintf("gemini-direct-batch-%d", input.ArticleSourceID)
		res.Articles = append(res.Articles, articleResult)
	}
	return res, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func acceptedArticleFromArtifact(artifact *geminidirect.ParsedArtifact) AcceptedArticle {
	article := AcceptedArticle{
		ArticleSourceID: artifact.ArticleSourceID,
		ArticleURL:      strings.TrimSpace(artifact.ArticleURL),
		YouTubeURL:      strings.TrimSpace(artifact.YouTubeURL),
		PresenterName:   artifact.PresenterName,
	}
	model := strings.TrimSpace(artifact.Model)
	var modelPtr *string
	if model != "" {
		modelPtr = &model
	}
	for _, spot := range artifact.Spots {
		evidence := strings.TrimSpace(spot.Evidence)
		if evidence == "" {
			evidence = strings.TrimSpace(spot.Notes)
		}
		var evidencePtr *string
		if evidence != "" {
			evidencePtr = &evidence
		}
		article.Spots = append(article.Spots, AcceptedSpot{
			ArticleSourceID:         artifact.ArticleSourceID,
			ArticleURL:              article.ArticleURL,
			YouTubeURL:              article.YouTubeURL,
			Place:                   strings.TrimSpace(spot.Place),
			Address:                 spot.Address,
			YouTubeTimestampSeconds: spot.YouTubeTimestampSeconds,
			Evidence:                evidencePtr,
			Confidence:              spot.Confidence,
			Model:                   modelPtr,
		})
	}
	return article
}

func missingInputs(input ArticleInput) []string {
	var missing []string
	if strings.TrimSpace(input.ArticleURL) == "" {
		missing = append(missing, "article URL")
	}
	if strings.TrimSpace(input.YouTubeURL) == "" {
		missing = append(missing, "YouTube URL")
	}
	return missing
}
