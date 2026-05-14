package repository

import (
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func seedExportFixture(t *testing.T, repo *Repository) {
	t.Helper()
	sourceID, err := repo.UpsertArticleSource("https://example.com/a")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	fetchID, err := repo.UpsertArticleFetch(sourceID, "<html>a</html>")
	if err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}
	audioID, err := repo.UpsertAudioSource(fetchID, "https://youtube.com/watch?v=abc", "mp3", "audio/mpeg", []byte("x"))
	if err != nil {
		t.Fatalf("UpsertAudioSource: %v", err)
	}
	trID, err := repo.UpsertAudioTranscription(models.ArticleAudioTranscription{AudioSourceID: audioID, Provider: "murmel", Language: "nl", HTTPStatus: 200, ResponseJSON: `{"segments":[]}`, ResponseByteSize: 14})
	if err != nil {
		t.Fatalf("UpsertAudioTranscription: %v", err)
	}
	ts := 15.13
	spotMentionID, err := repo.UpsertSpotMention(trID, "Stopera", &ts, &ts, &ts)
	if err != nil {
		t.Fatalf("UpsertSpotMention: %v", err)
	}
	placeID := "place_1"
	formatted := "Stopera Amsterdam"
	geoID, err := repo.UpsertSpotGoogleGeocode(spotMentionID, &placeID, 52.0, 4.0, &formatted, "ok")
	if err != nil {
		t.Fatalf("UpsertSpotGoogleGeocode: %v", err)
	}
	if err := repo.LinkArticleSpot(sourceID, geoID); err != nil {
		t.Fatalf("LinkArticleSpot: %v", err)
	}
	presenterID, err := repo.UpsertPresenter("Ray Fuego")
	if err != nil {
		t.Fatalf("UpsertPresenter: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter: %v", err)
	}
}

func TestExportDataIncludesRequiredFields(t *testing.T) {
	repo := newTestRepo(t)
	seedExportFixture(t, repo)

	data, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	if len(data.Spots) != 1 {
		t.Fatalf("expected 1 spot, got %d", len(data.Spots))
	}
	spot := data.Spots[0]
	if spot.PlaceID == "" || spot.SpotName == "" || spot.PresenterName == "" || spot.YouTubeLink == "" {
		t.Fatalf("expected required spot fields, got %+v", spot)
	}
	if spot.YouTubeLink != "https://youtu.be/abc?t=15" {
		t.Fatalf("expected timestamped youtube link, got %s", spot.YouTubeLink)
	}
	if len(data.Presenters) != 1 || data.Presenters[0].PresenterName != "Ray Fuego" {
		t.Fatalf("unexpected presenters: %+v", data.Presenters)
	}
}

func TestExportDataDeterministicOrdering(t *testing.T) {
	repo := newTestRepo(t)
	seedExportFixture(t, repo)

	sourceID, err := repo.UpsertArticleSource("https://example.com/b")
	if err != nil {
		t.Fatalf("UpsertArticleSource(b): %v", err)
	}
	fetchID, err := repo.UpsertArticleFetch(sourceID, "<html>b</html>")
	if err != nil {
		t.Fatalf("UpsertArticleFetch(sourceID=%d): %v", sourceID, err)
	}
	audioID, err := repo.UpsertAudioSource(fetchID, "https://youtube.com/watch?v=def", "mp3", "audio/mpeg", []byte("x"))
	if err != nil {
		t.Fatalf("UpsertAudioSource(fetchID=%d): %v", fetchID, err)
	}
	trID, err := repo.UpsertAudioTranscription(models.ArticleAudioTranscription{AudioSourceID: audioID, Provider: "murmel", Language: "nl", HTTPStatus: 200, ResponseJSON: `{"segments":[]}`, ResponseByteSize: 14})
	if err != nil {
		t.Fatalf("UpsertAudioTranscription(audioID=%d): %v", audioID, err)
	}
	ts2 := 44.0
	mentionID, err := repo.UpsertSpotMention(trID, "B spot", &ts2, &ts2, &ts2)
	if err != nil {
		t.Fatalf("UpsertSpotMention(transcriptionID=%d): %v", trID, err)
	}
	placeID := "a_place"
	formatted := "B spot Amsterdam"
	geoID, err := repo.UpsertSpotGoogleGeocode(mentionID, &placeID, 52.0, 4.0, &formatted, "ok")
	if err != nil {
		t.Fatalf("UpsertSpotGoogleGeocode(mentionID=%d): %v", mentionID, err)
	}
	if err := repo.LinkArticleSpot(sourceID, geoID); err != nil {
		t.Fatalf("LinkArticleSpot(sourceID=%d, geoID=%d): %v", sourceID, geoID, err)
	}
	pid, err := repo.UpsertPresenter("Alice")
	if err != nil {
		t.Fatalf("UpsertPresenter(Alice): %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, pid); err != nil {
		t.Fatalf("LinkArticlePresenter(sourceID=%d, presenterID=%d): %v", sourceID, pid, err)
	}

	data1, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData first: %v", err)
	}
	data2, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData second: %v", err)
	}
	if len(data1.Spots) < 2 || data1.Spots[0].PlaceID > data1.Spots[1].PlaceID {
		t.Fatalf("spots not sorted by placeId: %+v", data1.Spots)
	}
	if len(data1.Presenters) < 2 || data1.Presenters[0].PresenterName > data1.Presenters[1].PresenterName {
		t.Fatalf("presenters not sorted by presenterName: %+v", data1.Presenters)
	}
	if data1.Spots[0] != data2.Spots[0] || data1.Presenters[0] != data2.Presenters[0] {
		t.Fatalf("expected deterministic output across runs")
	}
}

func TestExportDataEmptyDataset(t *testing.T) {
	repo := newTestRepo(t)
	data, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	if len(data.Spots) != 0 || len(data.Presenters) != 0 {
		t.Fatalf("expected empty arrays, got %+v", data)
	}
}
