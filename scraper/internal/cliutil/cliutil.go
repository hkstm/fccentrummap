package cliutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DefaultDBPath() string {
	if fromEnv := strings.TrimSpace(os.Getenv("SPOTS_DB_PATH")); fromEnv != "" {
		return filepath.Clean(fromEnv)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean("../data/spots.db")
	}

	scraperRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	return filepath.Join(scraperRoot, "..", "data", "spots.db")
}

func SafeExt(ext string) string {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return "bin"
	}
	for _, r := range ext {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return "bin"
		}
	}
	return ext
}
