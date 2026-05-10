package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
)

func ReadJSON[T any](path string) (T, error) {
	var out T
	b, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, fmt.Errorf("invalid json in %s: %w", path, err)
	}
	return out, nil
}

func WriteStageArtifact(rootDir, stage, identity, payloadType string, payload any) (string, error) {
	baseDir := strings.TrimSpace(rootDir)
	if baseDir == "" {
		baseDir = cliutil.DefaultDataDir()
	}
	out := cliutil.StageArtifactPath(baseDir, stage, identity, payloadType, "json")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(out, b, 0o644); err != nil {
		return "", err
	}
	return out, nil
}

func IdentityFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(strings.TrimSpace(path)), filepath.Ext(strings.TrimSpace(path)))
	parts := strings.SplitN(base, "__", 2)
	return strings.TrimSpace(parts[0])
}

func IdentityFromURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return ""
	}
	s := strings.NewReplacer("https://", "", "http://", "", "/", "-", "?", "-", "&", "-", "=", "-", ":", "-").Replace(trimmed)
	s = strings.Trim(s, "-")
	if s == "" {
		return "collect-article-urls"
	}
	return s
}
