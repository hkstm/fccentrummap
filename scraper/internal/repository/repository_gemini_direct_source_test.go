package repository

import (
	"database/sql"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func TestInitSchemaCreatesGeminiDirectTablesConstraintsAndForeignKeys(t *testing.T) {
	repo := newTestRepo(t)
	for _, table := range []string{"gemini_direct_spot_mentions", "gemini_direct_spot_google_geocodes", "gemini_direct_article_spots", "gemini_direct_spot_corrections"} {
		var name string
		if err := repo.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
	}

	assertColumnExists(t, repo, "gemini_direct_spot_mentions", "confidence")
	assertColumnExists(t, repo, "gemini_direct_spot_mentions", "evidence")
	assertColumnExists(t, repo, "gemini_direct_spot_google_geocodes", "primary_type")

	sourceID, err := repo.UpsertArticleSource("https://example.com/gemini")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	ts := 12.0
	conf := 0.7
	evidence := "seen in video"
	model := "gemini-test"
	mentionID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, ArticleURL: "https://example.com/gemini", YouTubeURL: "https://youtube.com/watch?v=abc", Place: "Stopera", YouTubeTimestampSeconds: &ts, Evidence: &evidence, Confidence: &conf, Model: &model})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention: %v", err)
	}
	mentionID2, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, ArticleURL: "https://example.com/gemini", YouTubeURL: "https://youtube.com/watch?v=abc", Place: "Stopera", YouTubeTimestampSeconds: &ts, Evidence: &evidence, Confidence: &conf, Model: &model})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention duplicate: %v", err)
	}
	if mentionID != mentionID2 {
		t.Fatalf("duplicate Gemini mention changed logical id: %d vs %d", mentionID, mentionID2)
	}

	if _, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: 999999, YouTubeURL: "https://youtube.com/watch?v=abc", Place: "Bad"}); err == nil {
		t.Fatalf("expected FK/article validation failure for unknown article_source_id")
	}
	badTS := -1.0
	if _, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, YouTubeURL: "https://youtube.com/watch?v=abc", Place: "Bad TS", YouTubeTimestampSeconds: &badTS}); err == nil {
		t.Fatalf("expected timestamp constraint/validation failure")
	}
	badConfidence := 1.5
	if _, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, YouTubeURL: "https://youtube.com/watch?v=abc", Place: "Bad Confidence", Confidence: &badConfidence}); err == nil {
		t.Fatalf("expected confidence constraint/validation failure")
	}
}

func TestListGeminiDirectInputsFallsBackToFetchedArticleYouTubeURL(t *testing.T) {
	repo := newTestRepo(t)
	sourceID, err := repo.UpsertArticleSource("https://example.com/html-video")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(sourceID, `<html><meta property="article:published_time" content="2025-01-03T10:00:00+01:00"><iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ"></iframe></html>`); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}

	inputs, err := repo.ListGeminiDirectInputs()
	if err != nil {
		t.Fatalf("ListGeminiDirectInputs: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("got %d inputs, want 1: %+v", len(inputs), inputs)
	}
	if inputs[0].YouTubeURL != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Fatalf("YouTubeURL = %q", inputs[0].YouTubeURL)
	}
}

func TestGeminiDirectGeocodeAndExportIncludeValidMentions(t *testing.T) {
	repo := newTestRepo(t)
	sourceID, err := repo.UpsertArticleSource("https://example.com/confidence-filter")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(sourceID, `<html><meta property="article:published_time" content="2025-01-03T10:00:00+01:00"></html>`); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}
	oldPresenterID, err := repo.UpsertPresenter("Old Presenter")
	if err != nil {
		t.Fatalf("UpsertPresenter old: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, oldPresenterID); err != nil {
		t.Fatalf("LinkArticlePresenter old: %v", err)
	}
	newPresenterID, err := repo.UpsertPresenter("New Presenter")
	if err != nil {
		t.Fatalf("UpsertPresenter new: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, newPresenterID); err != nil {
		t.Fatalf("LinkArticlePresenter new: %v", err)
	}
	highConfidence := 0.9
	lowConfidence := 0.89
	highID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, ArticleURL: "https://example.com/confidence-filter", YouTubeURL: "https://youtube.com/watch?v=conf", Place: "High Confidence", Confidence: &highConfidence})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention high: %v", err)
	}
	lowID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, ArticleURL: "https://example.com/confidence-filter", YouTubeURL: "https://youtube.com/watch?v=conf", Place: "Low Confidence", Confidence: &lowConfidence})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention low: %v", err)
	}
	noConfidenceID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, ArticleURL: "https://example.com/confidence-filter", YouTubeURL: "https://youtube.com/watch?v=conf", Place: "No Confidence"})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention no confidence: %v", err)
	}

	mentions, err := repo.ListSpotMentionsWithoutGeocodeForSource(models.SpotSourceGeminiDirect)
	if err != nil {
		t.Fatalf("ListSpotMentionsWithoutGeocodeForSource: %v", err)
	}
	if len(mentions) != 3 || mentions[0].SpotMentionID != highID || mentions[1].SpotMentionID != lowID || mentions[2].SpotMentionID != noConfidenceID {
		t.Fatalf("Gemini geocode list should include all valid mentions, got %+v", mentions)
	}

	placeIDHigh := "high-place-id"
	placeIDLow := "low-place-id"
	placeIDNone := "none-place-id"
	formatted := "Amsterdam"
	if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(models.SpotSourceGeminiDirect, highID, &placeIDHigh, 52.1, 4.1, &formatted, "ok", sourceID, PlaceTypeMetadata{}); err != nil {
		t.Fatalf("geocode high: %v", err)
	}
	if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(models.SpotSourceGeminiDirect, lowID, &placeIDLow, 52.2, 4.2, &formatted, "ok", sourceID, PlaceTypeMetadata{}); err != nil {
		t.Fatalf("geocode low: %v", err)
	}
	if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(models.SpotSourceGeminiDirect, noConfidenceID, &placeIDNone, 52.3, 4.3, &formatted, "ok", sourceID, PlaceTypeMetadata{}); err != nil {
		t.Fatalf("geocode no confidence: %v", err)
	}
	data, err := repo.ExportDataForSource(models.SpotSourceGeminiDirect)
	if err != nil {
		t.Fatalf("ExportDataForSource: %v", err)
	}
	if len(data.Spots) != 3 || data.Spots[0].PresenterName != "New Presenter" {
		t.Fatalf("Gemini export should include all geocoded valid mentions with latest presenter, got %+v", data.Spots)
	}
}

func TestGeminiDirectExportOrdersSpotsChronologically(t *testing.T) {
	repo := newTestRepo(t)

	olderSourceID, err := repo.UpsertArticleSource("https://example.com/older")
	if err != nil {
		t.Fatalf("UpsertArticleSource older: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(olderSourceID, `<html><meta property="article:published_time" content="2025-01-01T10:00:00+01:00"></html>`); err != nil {
		t.Fatalf("UpsertArticleFetch older: %v", err)
	}
	newerSourceID, err := repo.UpsertArticleSource("https://example.com/newer")
	if err != nil {
		t.Fatalf("UpsertArticleSource newer: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(newerSourceID, `<html><meta property="article:published_time" content="2025-01-03T10:00:00+01:00"></html>`); err != nil {
		t.Fatalf("UpsertArticleFetch newer: %v", err)
	}
	presenterID, err := repo.UpsertPresenter("Presenter")
	if err != nil {
		t.Fatalf("UpsertPresenter: %v", err)
	}
	if err := repo.LinkArticlePresenter(olderSourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter older: %v", err)
	}
	if err := repo.LinkArticlePresenter(newerSourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter newer: %v", err)
	}

	newerMentionID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: newerSourceID, ArticleURL: "https://example.com/newer", YouTubeURL: "https://youtube.com/watch?v=new", Place: "Newer Spot"})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention newer: %v", err)
	}
	olderMentionID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: olderSourceID, ArticleURL: "https://example.com/older", YouTubeURL: "https://youtube.com/watch?v=old", Place: "Older Spot"})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention older: %v", err)
	}
	formatted := "Amsterdam"
	newerPlaceID := "newer-place"
	if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(models.SpotSourceGeminiDirect, newerMentionID, &newerPlaceID, 52.2, 4.2, &formatted, "ok", newerSourceID, PlaceTypeMetadata{}); err != nil {
		t.Fatalf("geocode newer: %v", err)
	}
	olderPlaceID := "older-place"
	if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(models.SpotSourceGeminiDirect, olderMentionID, &olderPlaceID, 52.1, 4.1, &formatted, "ok", olderSourceID, PlaceTypeMetadata{}); err != nil {
		t.Fatalf("geocode older: %v", err)
	}

	data, err := repo.ExportDataForSource(models.SpotSourceGeminiDirect)
	if err != nil {
		t.Fatalf("ExportDataForSource: %v", err)
	}
	if len(data.Spots) != 2 || data.Spots[0].ArticleURL != "https://example.com/older" || data.Spots[1].ArticleURL != "https://example.com/newer" {
		t.Fatalf("spots should export oldest-to-newest so newest is last/highest z-index: %+v", data.Spots)
	}
}

func TestSpotCategoryMappingsRepositoryAndExportFallback(t *testing.T) {
	repo := newTestRepo(t)
	for _, column := range []string{"primary_type_display_name", "category_name", "confidence", "reason", "model", "mapped_at"} {
		assertColumnExists(t, repo, "spot_category_mappings", column)
	}

	sourceID, err := repo.UpsertArticleSource("https://example.com/category")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(sourceID, `<html><meta property="article:published_time" content="2025-01-03T10:00:00+01:00"></html>`); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}
	presenterID, err := repo.UpsertPresenter("Presenter")
	if err != nil {
		t.Fatalf("UpsertPresenter: %v", err)
	}
	if err := repo.LinkArticlePresenter(sourceID, presenterID); err != nil {
		t.Fatalf("LinkArticlePresenter: %v", err)
	}
	mentionID, err := repo.UpsertGeminiDirectSpotMention(GeminiDirectSpotMentionInput{ArticleSourceID: sourceID, YouTubeURL: "https://youtube.com/watch?v=cat", Place: "Cafe"})
	if err != nil {
		t.Fatalf("UpsertGeminiDirectSpotMention: %v", err)
	}
	placeID := "place"
	primary := " Café "
	if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(models.SpotSourceGeminiDirect, mentionID, &placeID, 52.1, 4.1, nil, "ok", sourceID, PlaceTypeMetadata{PrimaryTypeDisplayName: &primary}); err != nil {
		t.Fatalf("geocode: %v", err)
	}

	names, err := repo.ListDistinctPrimaryTypeDisplayNames()
	if err != nil {
		t.Fatalf("ListDistinctPrimaryTypeDisplayNames: %v", err)
	}
	if len(names) != 1 || names[0] != "Café" {
		t.Fatalf("names = %+v, want [Café]", names)
	}
	data, err := repo.ExportDataForSource(models.SpotSourceGeminiDirect)
	if err != nil {
		t.Fatalf("ExportDataForSource fallback: %v", err)
	}
	if data.Spots[0].CategoryName != "Overig" {
		t.Fatalf("fallback category = %q", data.Spots[0].CategoryName)
	}

	confidence := 0.95
	reason := "coffee"
	if err := repo.ReplaceSpotCategoryMappings([]SpotCategoryMapping{{PrimaryTypeDisplayName: "Café", CategoryName: "Cafés & Bakkerijen", Confidence: &confidence, Reason: &reason, Model: "model"}}); err != nil {
		t.Fatalf("ReplaceSpotCategoryMappings: %v", err)
	}
	data, err = repo.ExportDataForSource(models.SpotSourceGeminiDirect)
	if err != nil {
		t.Fatalf("ExportDataForSource mapped: %v", err)
	}
	if data.Spots[0].CategoryName != "Cafés & Bakkerijen" || len(data.Categories) != 1 || data.Categories[0].CategoryName != "Cafés & Bakkerijen" {
		t.Fatalf("mapped export = %+v", data)
	}
	if err := repo.ReplaceSpotCategoryMappings(nil); err != nil {
		t.Fatalf("ReplaceSpotCategoryMappings clear: %v", err)
	}
	if got := countRows(t, repo.db, "spot_category_mappings"); got != 0 {
		t.Fatalf("mapping row count after replacement = %d", got)
	}
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func assertColumnExists(t *testing.T, repo *Repository, table, column string) {
	t.Helper()
	exists, err := repo.columnExists(table, column)
	if err != nil {
		t.Fatalf("columnExists %s.%s: %v", table, column, err)
	}
	if !exists {
		t.Fatalf("expected column %s.%s to exist", table, column)
	}
}
