package main

import "testing"

func TestValidateStageModeSupportsGeocodeSQLite(t *testing.T) {
	err := validateStageMode("geocode-spots", "sqlite")
	if err != nil {
		t.Fatalf("expected sqlite mode support for geocode-spots, got: %v", err)
	}
}

func TestValidateStageModeInvalidValue(t *testing.T) {
	err := validateStageMode("fetch-articles", "bogus")
	if err == nil {
		t.Fatalf("expected invalid io error")
	}
	if got := err.Error(); got == "" {
		t.Fatalf("expected actionable error message")
	}
}

func TestExperimentalCommandsAvailableAndModesExplicit(t *testing.T) {
	if got := extractSpotsGeminiDirectCommand().Name; got != "extract-spots-gemini-direct" {
		t.Fatalf("unexpected Gemini-direct command name %q", got)
	}
	if err := validateStageMode("extract-spots-gemini-direct", "sqlite"); err != nil {
		t.Fatalf("expected sqlite support: %v", err)
	}
	if err := validateStageMode("extract-spots-gemini-direct", "file"); err == nil {
		t.Fatalf("expected file mode to be unsupported")
	}
}

func TestExperimentalRequestNormalization(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "key")
	req, err := normalizeExtractSpotsGeminiDirectRequest(" db.sqlite ", "", "", true, 1)
	if err != nil {
		t.Fatalf("normalizeExtractSpotsGeminiDirectRequest: %v", err)
	}
	if req.DBPath != "db.sqlite" || req.OutDir == "" || req.Model == "" || req.APIKey != "key" || !req.Force || req.Limit != 1 {
		t.Fatalf("unexpected request normalization: %+v", req)
	}
}
