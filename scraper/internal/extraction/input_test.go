package extraction

import (
	"strings"
	"testing"
)

func TestParseSentenceUnitsSuccess(t *testing.T) {
	raw := `{"segments":[{"text":"Eerste zin","start":1.2},{"text":"Tweede zin","start":3.4}]}`
	units, err := ParseSentenceUnits(raw)
	if err != nil {
		t.Fatalf("ParseSentenceUnits() error = %v", err)
	}
	if len(units) != 2 {
		t.Fatalf("expected 2 units, got %d", len(units))
	}
	if units[0].Text != "Eerste zin" || units[0].Start != 1.2 {
		t.Fatalf("unexpected first unit: %+v", units[0])
	}
}

func TestParseSentenceUnitsRejectsFullTextOnly(t *testing.T) {
	raw := `{"text":"Alleen platte tekst zonder segmenten"}`
	_, err := ParseSentenceUnits(raw)
	if err == nil {
		t.Fatal("expected error for full-text-only payload")
	}
	if !strings.Contains(err.Error(), "full-text-only") {
		t.Fatalf("expected full-text-only error, got %v", err)
	}
}

func TestParseSentenceUnitsRejectsWordLevelOnly(t *testing.T) {
	raw := `{"words":[{"word":"Amsterdam","start":0.1}]}`
	_, err := ParseSentenceUnits(raw)
	if err == nil {
		t.Fatal("expected error for word-level-only payload")
	}
	if !strings.Contains(err.Error(), "word-level-only") {
		t.Fatalf("expected word-level-only error, got %v", err)
	}
}
