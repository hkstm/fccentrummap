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
