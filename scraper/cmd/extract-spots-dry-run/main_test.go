package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
)

func TestDryRunWritesTwoPassArtifactsAndDoesNotPersistExtractionRecord(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")

	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	url := "https://fccentrum.nl/story/de-spots-van-niels-oosthoek/"
	if err := repo.InsertArticleRaw(url, "<html><body>fallback html</body></html>", nil); err != nil {
		t.Fatalf("insert article raw: %v", err)
	}
	articleRaw, err := repo.GetArticleRawByURL(url)
	if err != nil || articleRaw == nil {
		t.Fatalf("get article raw: %+v err=%v", articleRaw, err)
	}

	if err := repo.ReplaceArticleTextExtraction(models.ArticleTextExtractionResult{
		ArticleRawID:   articleRaw.ArticleRawID,
		ExtractionMode: models.ArticleTextExtractionModeTrafilatura,
		Status:         models.ArticleTextExtractionStatusMatched,
		MatchedCount:   1,
		Contents: []models.ArticleTextContentInput{{
			SourceType: models.ArticleTextSourceTypeTrafilaturaText,
			Content:    "Niels tipt de Stopera en Oosterpark in Amsterdam.",
		}},
	}); err != nil {
		t.Fatalf("replace article text extraction: %v", err)
	}

	if err := repo.InsertArticleAudioSource(models.ArticleAudioSource{
		ArticleRawID: articleRaw.ArticleRawID,
		VideoID:      "video-1",
		YouTubeURL:   "https://youtube.com/watch?v=video-1",
		AudioFormat:  "mp3",
		MIMEType:     "audio/mpeg",
		AudioBlob:    []byte("audio"),
		ByteSize:     5,
	}); err != nil {
		t.Fatalf("insert audio source: %v", err)
	}
	audioSource, err := repo.GetArticleAudioSource(articleRaw.ArticleRawID)
	if err != nil || audioSource == nil {
		t.Fatalf("get audio source: %+v err=%v", audioSource, err)
	}

	transcriptionID, err := repo.UpsertArticleAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    audioSource.AudioSourceID,
		Provider:         "murmel",
		Language:         "nl",
		HTTPStatus:       200,
		ResponseJSON:     `{"segments":[{"text":"Vandaag starten we bij de Stopera","start":15.0},{"text":"Daarna gaan we naar Oosterpark","start":44.0}]}`,
		ResponseByteSize: 128,
	})
	if err != nil {
		t.Fatalf("upsert transcription: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		payload := buf.String()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(payload, "submit_refined_spots"):
			fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"functionCall":{"name":"submit_refined_spots","args":{"spots":[{"place":"Stopera","refinedSentenceStartTimestamp":12.0},{"place":"Oosterpark","refinedSentenceStartTimestamp":39.5}]}}}]}}]}`)
		default:
			fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"functionCall":{"name":"submit_spots","args":{"presenter_name":"Niels Oosthoek","spots":[{"place":"Stopera","sentenceStartTimestamp":15.0},{"place":"Oosterpark","sentenceStartTimestamp":44.0}]}}}]}}]}`)
		}
	}))
	defer server.Close()

	if err := runWithArgs([]string{
		"extract-spots-dry-run",
		"--db-path", dbPath,
		"--transcription-id", fmt.Sprintf("%d", transcriptionID),
		"--out-dir", tmpDir,
		"--gemma-model", "gemma-4-31b-it",
		"--gemini-api-key", "test-key",
		"--google-endpoint", server.URL,
	}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	paths := []string{
		"*_pass1_prompt.txt",
		"*_pass2_prompt.txt",
		"*_pass1_response.json",
		"*_pass2_response.json",
	}
	for _, p := range paths {
		matches, err := filepath.Glob(filepath.Join(tmpDir, p))
		if err != nil {
			t.Fatalf("glob %s: %v", p, err)
		}
		if len(matches) != 1 {
			t.Fatalf("expected exactly one artifact for %s, got %d (%v)", p, len(matches), matches)
		}
	}

	record, err := repo.GetLatestSpotExtractionRecord(articleRaw.ArticleRawID)
	if err != nil {
		t.Fatalf("GetLatestSpotExtractionRecord() error = %v", err)
	}
	if record != nil {
		t.Fatalf("expected dry-run to avoid DB persistence, got record %+v", record)
	}
}

func runWithArgs(args []string) error {
	oldArgs := os.Args
	oldFlagSet := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagSet
	}()

	os.Args = args
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	return run()
}
