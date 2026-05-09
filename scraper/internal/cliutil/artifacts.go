package cliutil

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var nonIdentityChars = regexp.MustCompile(`[^a-z0-9]+`)

func NormalizeIdentity(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = nonIdentityChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "unknown"
	}
	return s
}

func StageArtifactPath(baseDir, stage, identity, payloadType, ext string) string {
	id := NormalizeIdentity(identity)
	cleanStage := NormalizeIdentity(stage)
	cleanType := NormalizeIdentity(payloadType)
	cleanExt := strings.TrimPrefix(strings.TrimSpace(ext), ".")
	if cleanExt == "" {
		cleanExt = "json"
	}
	name := fmt.Sprintf("%s__%s__%s.%s", id, cleanStage, cleanType, cleanExt)
	return filepath.Join(baseDir, "stages", cleanStage, name)
}
