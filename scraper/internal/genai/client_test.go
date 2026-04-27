package genai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/extraction"
)

func TestGenerateContentFormatsRequest(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotContentType string
	var gotAPIKeyHeader string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.String()
		gotContentType = r.Header.Get("Content-Type")
		gotAPIKeyHeader = r.Header.Get("x-goog-api-key")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"functionCall":{"name":"submit_spots","args":{"spots":[]}}}]}}]}`))
	}))
	defer ts.Close()

	client := NewClientWithEndpoint("test-key", "gemma-4-31b-it", ts.URL)
	_, err := client.GenerateContent(context.Background(), "hallo prompt", extraction.GenerateContentConfig())
	if err != nil {
		t.Fatalf("GenerateContent() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/v1beta/models/gemma-4-31b-it:generateContent") {
		t.Fatalf("unexpected request path: %s", gotPath)
	}
	if !strings.Contains(strings.ToLower(gotContentType), "application/json") {
		t.Fatalf("expected application/json content type, got %q", gotContentType)
	}
	if gotAPIKeyHeader != "test-key" {
		t.Fatalf("expected x-goog-api-key header to be test-key, got %q", gotAPIKeyHeader)
	}

	tools, ok := gotBody["tools"].([]any)
	if !ok || len(tools) == 0 {
		t.Fatalf("request body missing tools: %+v", gotBody)
	}
	tool0, ok := tools[0].(map[string]any)
	if !ok {
		t.Fatalf("request body tools[0] has unexpected type: %+v", gotBody)
	}
	decls, ok := tool0["functionDeclarations"].([]any)
	if !ok || len(decls) == 0 {
		t.Fatalf("request body missing functionDeclarations: %+v", gotBody)
	}
	decl0, ok := decls[0].(map[string]any)
	if !ok {
		t.Fatalf("request body functionDeclarations[0] has unexpected type: %+v", gotBody)
	}
	if decl0["name"] != "submit_spots" {
		t.Fatalf("expected function name submit_spots, got %v (body=%+v)", decl0["name"], gotBody)
	}
	toolCfg, ok := gotBody["toolConfig"].(map[string]any)
	if !ok {
		t.Fatalf("request body missing toolConfig: %+v", gotBody)
	}
	fc, ok := toolCfg["functionCallingConfig"].(map[string]any)
	if !ok || fc["mode"] != "ANY" {
		t.Fatalf("request body missing function calling mode ANY: %+v", gotBody)
	}
}

func TestGenerateContentNon2xxIncludesDiagnostics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer ts.Close()

	client := NewClientWithEndpoint("test-key", "gemma-4-31b-it", ts.URL)
	_, err := client.GenerateContent(context.Background(), "hallo", extraction.GenerateContentConfig())
	if err == nil {
		t.Fatal("expected non-2xx error")
	}
	if !strings.Contains(err.Error(), "400") || !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("expected status and body diagnostics, got %v", err)
	}
}
