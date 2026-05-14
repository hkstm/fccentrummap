package contracttests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hkstm/fccentrummap/internal/pipeline/acquireaudio"
	"github.com/hkstm/fccentrummap/internal/pipeline/collectarticleurls"
	"github.com/hkstm/fccentrummap/internal/pipeline/exportdata"
	"github.com/hkstm/fccentrummap/internal/pipeline/extractspots"
	"github.com/hkstm/fccentrummap/internal/pipeline/fetcharticles"
	"github.com/hkstm/fccentrummap/internal/pipeline/geocodespots"
	"github.com/hkstm/fccentrummap/internal/pipeline/transcribeaudio"
)

func writeJSON(t *testing.T, dir, name string, v any) string {
	t.Helper()
	p := filepath.Join(dir, name)
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %s: %v", name, err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFileAdaptersContractValidationAndOutput(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := collectarticleurls.NewFileAdapter().Run(ctx, collectarticleurls.Request{InputPath: writeJSON(t, dir, "collect.json", map[string]any{"urls": []string{"https://a"}})}); err == nil {
		t.Fatalf("expected collectarticleurls file adapter to return not implemented")
	}
	if _, err := fetcharticles.NewFileAdapter().Run(ctx, fetcharticles.Request{InputPath: writeJSON(t, dir, "fetch.json", map[string]any{"articleUrls": []string{"https://a"}})}); err == nil {
		t.Fatalf("expected fetcharticles file adapter to return not implemented")
	}
	if _, err := acquireaudio.NewFileAdapter().Run(ctx, acquireaudio.Request{InputPath: writeJSON(t, dir, "acquire.json", map[string]any{"articles": []map[string]any{{"url": "https://a", "videoId": "vid"}}})}); err == nil {
		t.Fatalf("expected acquireaudio file adapter to return not implemented")
	}
	if _, err := transcribeaudio.NewFileAdapter().Run(ctx, transcribeaudio.Request{InputPath: writeJSON(t, dir, "transcribe.json", map[string]any{"audioSourceId": 1, "audioBlobBase64": "Zm9v", "language": "nl"})}); err == nil {
		t.Fatalf("expected transcribeaudio file adapter to return not implemented")
	}
	if _, err := extractspots.NewFileAdapter().Run(ctx, extractspots.Request{InputPath: writeJSON(t, dir, "extract.json", map[string]any{"transcriptionJson": "{}", "articleText": "abc"})}); err == nil {
		t.Fatalf("expected extractspots file adapter to return not implemented")
	}
	if _, err := exportdata.NewFileAdapter().Run(ctx, exportdata.Request{InputPath: writeJSON(t, dir, "export.json", map[string]any{"authors": []string{"a"}})}); err == nil {
		t.Fatalf("expected exportdata file adapter to return not implemented")
	}

	t.Setenv("PRODUCTION_GOOGLE_MAPS_API_KEY", "")
	if _, err := geocodespots.NewFileAdapter().Run(ctx, geocodespots.Request{InputPath: writeJSON(t, dir, "geocode.json", map[string]any{"query": "Amsterdam"})}); err == nil {
		t.Fatalf("expected geocodespots file adapter to fail without PRODUCTION_GOOGLE_MAPS_API_KEY")
	}
}
