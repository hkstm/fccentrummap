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

	collectResp, err := collectarticleurls.NewFileAdapter().Run(ctx, collectarticleurls.Request{Identity: "id1", InputPath: writeJSON(t, dir, "collect.json", map[string]any{"urls": []string{"https://a"}})})
	if err != nil {
		t.Fatalf("collectarticleurls file adapter: %v", err)
	}
	if collectResp.Identity != "id1" || collectResp.Stage != "collectarticleurls" || collectResp.OutputPath == "" {
		t.Fatalf("collectarticleurls contract mismatch: %+v", collectResp)
	}
	if _, err := os.Stat(collectResp.OutputPath); err != nil {
		t.Fatalf("collectarticleurls output missing: %v", err)
	}

	fetchResp, err := fetcharticles.NewFileAdapter().Run(ctx, fetcharticles.Request{Identity: "id2", InputPath: writeJSON(t, dir, "fetch.json", map[string]any{"articleUrls": []string{"https://a"}})})
	if err != nil {
		t.Fatalf("fetcharticles file adapter: %v", err)
	}
	if fetchResp.Identity != "id2" || fetchResp.Stage != "fetcharticles" || fetchResp.OutputPath == "" {
		t.Fatalf("fetcharticles contract mismatch: %+v", fetchResp)
	}
	if _, err := os.Stat(fetchResp.OutputPath); err != nil {
		t.Fatalf("fetcharticles output missing: %v", err)
	}

	acquireResp, err := acquireaudio.NewFileAdapter().Run(ctx, acquireaudio.Request{Identity: "id3", InputPath: writeJSON(t, dir, "acquire.json", map[string]any{"articles": []map[string]any{{"url": "https://a", "videoId": "vid"}}})})
	if err != nil {
		t.Fatalf("acquireaudio file adapter: %v", err)
	}
	if acquireResp.Identity != "id3" || acquireResp.Stage != "acquireaudio" || acquireResp.OutputPath == "" {
		t.Fatalf("acquireaudio contract mismatch: %+v", acquireResp)
	}
	if _, err := os.Stat(acquireResp.OutputPath); err != nil {
		t.Fatalf("acquireaudio output missing: %v", err)
	}

	transcribeResp, err := transcribeaudio.NewFileAdapter().Run(ctx, transcribeaudio.Request{Identity: "id4", InputPath: writeJSON(t, dir, "transcribe.json", map[string]any{"audioSourceId": 1, "audioBlobBase64": "Zm9v", "language": "nl"})})
	if err != nil {
		t.Fatalf("transcribeaudio file adapter: %v", err)
	}
	if transcribeResp.Identity != "id4" || transcribeResp.Stage != "transcribeaudio" || transcribeResp.OutputPath == "" {
		t.Fatalf("transcribeaudio contract mismatch: %+v", transcribeResp)
	}
	if _, err := os.Stat(transcribeResp.OutputPath); err != nil {
		t.Fatalf("transcribeaudio output missing: %v", err)
	}

	extractResp, err := extractspots.NewFileAdapter().Run(ctx, extractspots.Request{Identity: "id5", InputPath: writeJSON(t, dir, "extract.json", map[string]any{"transcriptionJson": "{}", "articleText": "abc"})})
	if err != nil {
		t.Fatalf("extractspots file adapter: %v", err)
	}
	if extractResp.Identity != "id5" || extractResp.Stage != "extractspots" || extractResp.OutputPath == "" {
		t.Fatalf("extractspots contract mismatch: %+v", extractResp)
	}
	if _, err := os.Stat(extractResp.OutputPath); err != nil {
		t.Fatalf("extractspots output missing: %v", err)
	}

	exportResp, err := exportdata.NewFileAdapter().Run(ctx, exportdata.Request{Identity: "id6", InputPath: writeJSON(t, dir, "export.json", map[string]any{"authors": []string{"a"}})})
	if err != nil {
		t.Fatalf("exportdata file adapter: %v", err)
	}
	if exportResp.Identity != "id6" || exportResp.Stage != "exportdata" || exportResp.OutputPath == "" {
		t.Fatalf("exportdata contract mismatch: %+v", exportResp)
	}
	if _, err := os.Stat(exportResp.OutputPath); err != nil {
		t.Fatalf("exportdata output missing: %v", err)
	}

	t.Setenv("GOOGLE_MAPS_API_KEY", "")
	if _, err := geocodespots.NewFileAdapter().Run(ctx, geocodespots.Request{Identity: "id7", InputPath: writeJSON(t, dir, "geocode.json", map[string]any{"query": "Amsterdam"})}); err == nil {
		t.Fatalf("expected geocodespots file adapter to fail without GOOGLE_MAPS_API_KEY")
	}
}
