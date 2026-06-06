package cliutil

import "testing"

func TestStageArtifactPathDeterministic(t *testing.T) {
	p1 := StageArtifactPath("/tmp/data", "extract-spots-gemini-direct", "Casa del Gusto Amsterdam", "candidates", "json")
	p2 := StageArtifactPath("/tmp/data", "extract-spots-gemini-direct", "Casa del Gusto Amsterdam", "candidates", "json")
	if p1 != p2 {
		t.Fatalf("expected deterministic artifact path, got %q != %q", p1, p2)
	}
	want := "/tmp/data/stages/extract-spots-gemini-direct/casa-del-gusto-amsterdam__extract-spots-gemini-direct__candidates.json"
	if p1 != want {
		t.Fatalf("unexpected path\nwant: %s\ngot:  %s", want, p1)
	}
}
