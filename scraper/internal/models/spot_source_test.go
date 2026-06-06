package models

import "testing"

func TestNormalizeSpotSourceDefaultsAndValidates(t *testing.T) {
	got, err := NormalizeSpotSource("")
	if err != nil || got != SpotSourceGeminiDirect {
		t.Fatalf("empty source = %q,%v want %q,nil", got, err, SpotSourceGeminiDirect)
	}
	if _, err := NormalizeSpotSource("legacy"); err == nil {
		t.Fatalf("expected unsupported source error")
	}
	if _, err := NormalizeSpotSource("other"); err == nil {
		t.Fatalf("expected unsupported source error")
	}
}
