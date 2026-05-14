package repository

import (
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func TestInitSchemaCreatesV2TablesAndConstraints(t *testing.T) {
	repo := newTestRepo(t)

	tables := []string{
		"article_sources", "article_fetches", "article_texts", "audio_sources", "audio_transcriptions",
		"spot_mentions", "spot_google_geocodes", "presenters", "article_presenters", "article_spots",
	}
	for _, table := range tables {
		var name string
		if err := repo.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
	}

	sourceID, err := repo.UpsertArticleSource("https://example.com/a")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(sourceID, "<html>a</html>"); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(999999, "<html>bad fk</html>"); err == nil {
		t.Fatalf("expected FK violation when upserting fetch for unknown article_source_id")
	}
}

func TestArticleFetchAndTextUpsertSemantics(t *testing.T) {
	repo := newTestRepo(t)
	sourceID, err := repo.UpsertArticleSource("https://example.com/b")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}

	fetchID1, err := repo.UpsertArticleFetch(sourceID, "<html>v1</html>")
	if err != nil {
		t.Fatalf("UpsertArticleFetch first: %v", err)
	}
	fetchID2, err := repo.UpsertArticleFetch(sourceID, "<html>v2</html>")
	if err != nil {
		t.Fatalf("UpsertArticleFetch second: %v", err)
	}
	if fetchID1 != fetchID2 {
		t.Fatalf("expected latest-only fetch row id to remain stable, got %d and %d", fetchID1, fetchID2)
	}

	textID1, err := repo.UpsertArticleText(fetchID1, "cleaned text v1")
	if err != nil {
		t.Fatalf("UpsertArticleText first: %v", err)
	}
	textID2, err := repo.UpsertArticleText(fetchID1, "cleaned text v2")
	if err != nil {
		t.Fatalf("UpsertArticleText second: %v", err)
	}
	if textID1 != textID2 {
		t.Fatalf("expected single-row text per fetch, got %d and %d", textID1, textID2)
	}
}

func TestExtractSpotsPresenterMaterialization(t *testing.T) {
	repo := newTestRepo(t)
	sourceID, err := repo.UpsertArticleSource("https://example.com/c")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	presenterID, err := repo.UpsertPresenter("Alice")
	if err != nil {
		t.Fatalf("UpsertPresenter: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter first: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter second: %v", err)
	}

	var n int
	if err := repo.db.QueryRow(`SELECT COUNT(*) FROM article_presenters WHERE article_source_id=? AND presenter_id=?`, sourceID, presenterID).Scan(&n); err != nil {
		t.Fatalf("count article_presenters: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected single presenter link row, got %d", n)
	}
}

func TestGeocodeAndArticleSpotLinking(t *testing.T) {
	repo := newTestRepo(t)
	sourceID, err := repo.UpsertArticleSource("https://example.com/d")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	fetchID, err := repo.UpsertArticleFetch(sourceID, "<html>d</html>")
	if err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}
	audioID, err := repo.UpsertAudioSource(fetchID, "https://youtube.com/watch?v=abc", "mp3", "audio/mpeg", []byte("x"))
	if err != nil {
		t.Fatalf("UpsertAudioSource: %v", err)
	}
	transcriptionID, err := repo.UpsertAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    audioID,
		Provider:         "murmel",
		Language:         "nl",
		HTTPStatus:       200,
		ResponseJSON:     `{"segments":[]}`,
		ResponseByteSize: 14,
	})
	if err != nil {
		t.Fatalf("UpsertAudioTranscription: %v", err)
	}
	spotMentionID, err := repo.UpsertSpotMention(transcriptionID, "Stopera", nil, nil, nil)
	if err != nil {
		t.Fatalf("UpsertSpotMention: %v", err)
	}
	formatted := "Stopera, Amsterdam"
	geoID, err := repo.UpsertSpotGoogleGeocode(spotMentionID, nil, 52.0, 4.0, &formatted, "ok")
	if err != nil {
		t.Fatalf("UpsertSpotGoogleGeocode: %v", err)
	}
	if err := repo.LinkArticleSpot(sourceID, geoID); err != nil {
		t.Fatalf("LinkArticleSpot: %v", err)
	}
	if err := repo.LinkArticleSpot(sourceID, geoID); err != nil {
		t.Fatalf("LinkArticleSpot second: %v", err)
	}
	var n int
	if err := repo.db.QueryRow(`SELECT COUNT(*) FROM article_spots WHERE article_source_id=? AND spot_google_geocode_id=?`, sourceID, geoID).Scan(&n); err != nil {
		t.Fatalf("count article_spots: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected single article_spots row, got %d", n)
	}
}
