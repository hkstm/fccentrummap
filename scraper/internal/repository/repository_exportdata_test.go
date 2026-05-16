package repository

import (
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func publishedMetaHTML(publishedAt string) string {
	return `<html><head><meta property="article:published_time" content="` + publishedAt + `"></head><body>article</body></html>`
}

func publishedJSONLDHTML(publishedAt string) string {
	return `<html><head><script type="application/ld+json">{"@type":"NewsArticle","datePublished":"` + publishedAt + `"}</script></head><body>article</body></html>`
}

func seedExportArticle(t *testing.T, repo *Repository, articleURL, presenterName, spotName, placeID, youtubeID, publishedHTML string) {
	t.Helper()
	sourceID, err := repo.UpsertArticleSource(articleURL)
	if err != nil {
		t.Fatalf("UpsertArticleSource(%s): %v", articleURL, err)
	}
	fetchID, err := repo.UpsertArticleFetch(sourceID, publishedHTML)
	if err != nil {
		t.Fatalf("UpsertArticleFetch(sourceID=%d): %v", sourceID, err)
	}
	audioID, err := repo.UpsertAudioSource(fetchID, "https://youtube.com/watch?v="+youtubeID, "mp3", "audio/mpeg", []byte("x"))
	if err != nil {
		t.Fatalf("UpsertAudioSource(fetchID=%d): %v", fetchID, err)
	}
	trID, err := repo.UpsertAudioTranscription(models.ArticleAudioTranscription{AudioSourceID: audioID, Provider: "murmel", Language: "nl", HTTPStatus: 200, ResponseJSON: `{"segments":[]}`, ResponseByteSize: 14})
	if err != nil {
		t.Fatalf("UpsertAudioTranscription(audioID=%d): %v", audioID, err)
	}
	ts := 15.13
	spotMentionID, err := repo.UpsertSpotMention(trID, spotName, &ts, &ts, &ts)
	if err != nil {
		t.Fatalf("UpsertSpotMention(transcriptionID=%d): %v", trID, err)
	}
	formatted := spotName + " Amsterdam"
	geoID, err := repo.UpsertSpotGoogleGeocode(spotMentionID, &placeID, 52.0, 4.0, &formatted, "ok")
	if err != nil {
		t.Fatalf("UpsertSpotGoogleGeocode(mentionID=%d): %v", spotMentionID, err)
	}
	if err := repo.LinkArticleSpot(sourceID, geoID); err != nil {
		t.Fatalf("LinkArticleSpot(sourceID=%d, geoID=%d): %v", sourceID, geoID, err)
	}
	presenterID, err := repo.UpsertPresenter(presenterName)
	if err != nil {
		t.Fatalf("UpsertPresenter(%s): %v", presenterName, err)
	}
	if err := repo.LinkArticlePresenter(sourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter(sourceID=%d, presenterID=%d): %v", sourceID, presenterID, err)
	}
}

func seedExportFixture(t *testing.T, repo *Repository) {
	t.Helper()
	seedExportArticle(t, repo, "https://example.com/a", "Ray Fuego", "Stopera", "place_1", "abc", publishedMetaHTML("2025-01-03T10:00:00+01:00"))
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
	if spot.PlaceID == "" || spot.SpotName == "" || spot.PresenterName == "" || spot.YouTubeLink == "" || spot.ArticleURL == "" {
		t.Fatalf("expected required spot fields, got %+v", spot)
	}
	if spot.Latitude != 52.0 || spot.Longitude != 4.0 {
		t.Fatalf("expected coordinates in export, got lat=%v lng=%v", spot.Latitude, spot.Longitude)
	}
	if spot.YouTubeLink != "https://youtu.be/abc?t=15" {
		t.Fatalf("expected timestamped youtube link, got %s", spot.YouTubeLink)
	}
	if spot.ArticleURL != "https://example.com/a" {
		t.Fatalf("expected article URL in export, got %s", spot.ArticleURL)
	}
	if len(data.Presenters) != 1 || data.Presenters[0].PresenterName != "Ray Fuego" {
		t.Fatalf("unexpected presenters: %+v", data.Presenters)
	}
}

func TestExportDataDeterministicOrdering(t *testing.T) {
	repo := newTestRepo(t)
	seedExportArticle(t, repo, "https://example.com/a", "Ray Fuego", "Stopera", "z_place", "abc", publishedMetaHTML("2025-01-03T10:00:00+01:00"))
	seedExportArticle(t, repo, "https://example.com/b", "Alice", "B spot", "a_place", "def", publishedMetaHTML("2025-01-02T10:00:00+01:00"))

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
	if len(data1.Presenters) != 2 || data1.Presenters[0].PresenterName != "Ray Fuego" || data1.Presenters[1].PresenterName != "Alice" {
		t.Fatalf("presenters not sorted by latest article published_at: %+v", data1.Presenters)
	}
	if data1.Spots[0] != data2.Spots[0] || data1.Presenters[0] != data2.Presenters[0] {
		t.Fatalf("expected deterministic output across runs")
	}
}

func TestExportDataOrdersPresentersByLatestArticlePublication(t *testing.T) {
	repo := newTestRepo(t)
	seedExportArticle(t, repo, "https://example.com/older", "Older Presenter", "Older spot", "place_older", "old", publishedMetaHTML("2025-01-01T10:00:00+01:00"))
	seedExportArticle(t, repo, "https://example.com/newer", "Newer Presenter", "Newer spot", "place_newer", "new", publishedMetaHTML("2025-01-05T10:00:00+01:00"))
	seedExportArticle(t, repo, "https://example.com/middle", "Middle Presenter", "Middle spot", "place_middle", "mid", publishedMetaHTML("2025-01-03T10:00:00+01:00"))

	data, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	got := []string{data.Presenters[0].PresenterName, data.Presenters[1].PresenterName, data.Presenters[2].PresenterName}
	want := []string{"Newer Presenter", "Middle Presenter", "Older Presenter"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("presenter order = %+v, want %+v", got, want)
		}
	}
}

func TestExportDataUsesPresenterNameTieBreakerForEqualPublicationTime(t *testing.T) {
	repo := newTestRepo(t)
	sameTime := "2025-01-03T10:00:00+01:00"
	seedExportArticle(t, repo, "https://example.com/z", "Zed", "Z spot", "place_z", "zed", publishedMetaHTML(sameTime))
	seedExportArticle(t, repo, "https://example.com/a", "Alice", "A spot", "place_a", "ali", publishedMetaHTML(sameTime))

	data, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	if len(data.Presenters) != 2 || data.Presenters[0].PresenterName != "Alice" || data.Presenters[1].PresenterName != "Zed" {
		t.Fatalf("presenters not sorted by tie-breaker: %+v", data.Presenters)
	}
}

func TestExportDataUsesLatestArticleForPresenter(t *testing.T) {
	repo := newTestRepo(t)
	seedExportArticle(t, repo, "https://example.com/shared-old", "Shared Presenter", "Old shared spot", "place_shared_old", "sho", publishedMetaHTML("2025-01-01T10:00:00+01:00"))
	seedExportArticle(t, repo, "https://example.com/other", "Other Presenter", "Other spot", "place_other", "oth", publishedMetaHTML("2025-01-03T10:00:00+01:00"))
	seedExportArticle(t, repo, "https://example.com/shared-new", "Shared Presenter", "New shared spot", "place_shared_new", "shn", publishedMetaHTML("2025-01-05T10:00:00+01:00"))

	data, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	if len(data.Presenters) != 2 || data.Presenters[0].PresenterName != "Shared Presenter" || data.Presenters[1].PresenterName != "Other Presenter" {
		t.Fatalf("presenters not sorted by latest presenter article: %+v", data.Presenters)
	}
}

func TestExportDataFailsWhenPublicationTimeMissing(t *testing.T) {
	repo := newTestRepo(t)
	seedExportArticle(t, repo, "https://example.com/missing", "No Date", "Missing date spot", "place_missing", "nod", "<html>no metadata</html>")

	_, err := repo.ExportData()
	if err == nil {
		t.Fatalf("expected ExportData to fail")
	}
	if !strings.Contains(err.Error(), "no stored publication time") {
		t.Fatalf("expected missing publication time diagnostic, got %v", err)
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
