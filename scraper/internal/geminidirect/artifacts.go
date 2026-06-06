package geminidirect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ArtifactPaths struct {
	Prompt string
	Raw    string
	Parsed string
}

func PathsForArticle(outDir string, articleSourceID int64) ArtifactPaths {
	base := fmt.Sprintf("article_%d", articleSourceID)
	return ArtifactPaths{
		Prompt: filepath.Join(outDir, base+"_prompt.txt"),
		Raw:    filepath.Join(outDir, base+"_raw_response.json"),
		Parsed: filepath.Join(outDir, base+"_parsed.json"),
	}
}

type ArtifactWriter struct {
	OutDir string
}

func (w ArtifactWriter) WritePrompt(articleSourceID int64, prompt string) (ArtifactPaths, error) {
	paths := PathsForArticle(w.OutDir, articleSourceID)
	if err := os.MkdirAll(w.OutDir, 0o755); err != nil {
		return paths, fmt.Errorf("create Gemini-direct artifact directory %s: %w", w.OutDir, err)
	}
	if err := os.WriteFile(paths.Prompt, []byte(prompt), 0o644); err != nil {
		return paths, fmt.Errorf("write Gemini-direct prompt artifact: %w", err)
	}
	return paths, nil
}

func (w ArtifactWriter) WriteRaw(articleSourceID int64, raw []byte) (ArtifactPaths, error) {
	paths := PathsForArticle(w.OutDir, articleSourceID)
	if err := os.MkdirAll(w.OutDir, 0o755); err != nil {
		return paths, fmt.Errorf("create Gemini-direct artifact directory %s: %w", w.OutDir, err)
	}
	if err := os.WriteFile(paths.Raw, raw, 0o644); err != nil {
		return paths, fmt.Errorf("write Gemini-direct raw response artifact: %w", err)
	}
	return paths, nil
}

func (w ArtifactWriter) WriteParsed(articleSourceID int64, artifact *ParsedArtifact) (ArtifactPaths, error) {
	paths := PathsForArticle(w.OutDir, articleSourceID)
	if err := os.MkdirAll(w.OutDir, 0o755); err != nil {
		return paths, fmt.Errorf("create Gemini-direct artifact directory %s: %w", w.OutDir, err)
	}
	body, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return paths, fmt.Errorf("marshal Gemini-direct parsed artifact: %w", err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(paths.Parsed, body, 0o644); err != nil {
		return paths, fmt.Errorf("write Gemini-direct parsed artifact: %w", err)
	}
	return paths, nil
}
