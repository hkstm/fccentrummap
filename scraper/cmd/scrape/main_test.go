package main

import (
	"os"
	"strings"
	"testing"
)

func TestValidateStageMode(t *testing.T) {
	if err := validateStageMode("collect-article-urls", "sqlite"); err != nil {
		t.Fatalf("expected supported mode, got %v", err)
	}
	if err := validateStageMode("geocode-spots", "sqlite"); err == nil {
		t.Fatalf("expected unsupported mode error")
	} else if !strings.Contains(err.Error(), "use --io file --in <path>") {
		t.Fatalf("expected actionable guidance, got %v", err)
	}
	if err := validateStageMode("init", "file"); err == nil {
		t.Fatalf("expected unsupported mode for init")
	}
	if err := validateStageMode("fetch-articles", "bogus"); err == nil {
		t.Fatalf("expected invalid io mode error")
	}
}

func TestValidateRequiredEnv(t *testing.T) {
	t.Setenv("MURMEL_API_KEY", "")
	t.Setenv("GOOGLE_MAPS_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GOOGLE_GENERATIVE_LANGUAGE_API_KEY", "")

	if err := validateRequiredEnv(); err == nil {
		t.Fatalf("expected error when required env vars are missing")
	}

	t.Setenv("MURMEL_API_KEY", "x")
	t.Setenv("GOOGLE_MAPS_API_KEY", "y")
	t.Setenv("GOOGLE_API_KEY", "z")
	if err := validateRequiredEnv(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	_ = os.Unsetenv("GOOGLE_API_KEY")
}
