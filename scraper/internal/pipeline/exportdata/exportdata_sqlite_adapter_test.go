package exportdata

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
)

func seedDBForExport(t *testing.T, dbPath string) {
	t.Helper()
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("repository.New: %v", err)
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	sourceID, err := repo.UpsertArticleSource("https://example.com/a")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	fetchID, err := repo.UpsertArticleFetch(sourceID, "<html>a</html>")
	if err != nil {
		t.Fatalf("UpsertArticleFetch(sourceID=%d): %v", sourceID, err)
	}
	audioID, err := repo.UpsertAudioSource(fetchID, "https://youtube.com/watch?v=abc", "mp3", "audio/mpeg", []byte("x"))
	if err != nil {
		t.Fatalf("UpsertAudioSource(fetchID=%d): %v", fetchID, err)
	}
	trID, err := repo.UpsertAudioTranscription(models.ArticleAudioTranscription{AudioSourceID: audioID, Provider: "murmel", Language: "nl", HTTPStatus: 200, ResponseJSON: `{"segments":[]}`, ResponseByteSize: 14})
	if err != nil {
		t.Fatalf("UpsertAudioTranscription(audioID=%d): %v", audioID, err)
	}
	ts := 15.13
	mentionID, err := repo.UpsertSpotMention(trID, "Stopera", &ts, &ts, &ts)
	if err != nil {
		t.Fatalf("UpsertSpotMention(transcriptionID=%d): %v", trID, err)
	}
	placeID := "place_1"
	formatted := "Stopera Amsterdam"
	geoID, err := repo.UpsertSpotGoogleGeocode(mentionID, &placeID, 52.0, 4.0, &formatted, "ok")
	if err != nil {
		t.Fatalf("UpsertSpotGoogleGeocode(mentionID=%d): %v", mentionID, err)
	}
	if err := repo.LinkArticleSpot(sourceID, geoID); err != nil {
		t.Fatalf("LinkArticleSpot(sourceID=%d, geoID=%d): %v", sourceID, geoID, err)
	}
	pid, err := repo.UpsertPresenter("Ray Fuego")
	if err != nil {
		t.Fatalf("UpsertPresenter: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, pid); err != nil {
		t.Fatalf("LinkArticlePresenter(sourceID=%d, presenterID=%d): %v", sourceID, pid, err)
	}
}

func TestSQLiteAdapterWritesJSONToCustomOutputPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	seedDBForExport(t, dbPath)

	outPath := filepath.Join(dir, "custom", "export.json")
	res, err := NewSQLiteAdapter().Run(context.Background(), Request{DBPath: dbPath, OutputPath: outPath})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.OutputPath != outPath {
		t.Fatalf("unexpected output path: %s", res.OutputPath)
	}
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(b)
	for _, required := range []string{"\"spots\"", "\"presenters\"", "\"placeId\"", "\"spotName\"", "\"presenterName\"", "\"latitude\"", "\"longitude\"", "\"youtubeLink\"", "\"articleUrl\""} {
		if !strings.Contains(got, required) {
			t.Fatalf("expected %s in output: %s", required, got)
		}
	}
}
