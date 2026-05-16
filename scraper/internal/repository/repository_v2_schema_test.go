package repository

import (
	"path/filepath"
	"strings"
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
	assertColumnExists(t, repo, "article_sources", "published_at")

	if _, err := repo.UpsertArticleFetch(sourceID, "<html>a</html>"); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(999999, "<html>bad fk</html>"); err == nil {
		t.Fatalf("expected FK violation when upserting fetch for unknown article_source_id")
	}
}

func assertColumnExists(t *testing.T, repo *Repository, table, column string) {
	t.Helper()
	rows, err := repo.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid       int
			name      string
			colType   string
			notNull   int
			defaultV  any
			primaryKY int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &primaryKY); err != nil {
			t.Fatalf("scan table info: %v", err)
		}
		if name == column {
			return
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate table info: %v", err)
	}
	t.Fatalf("expected column %s.%s to exist", table, column)
}

func TestUpsertArticleFetchStoresPublicationTime(t *testing.T) {
	repo := newTestRepo(t)
	sourceID, err := repo.UpsertArticleSource("https://example.com/published")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(sourceID, publishedMetaHTML("2025-01-03T10:00:00+01:00")); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}

	var stored string
	if err := repo.db.QueryRow(`SELECT published_at FROM article_sources WHERE article_source_id=?`, sourceID).Scan(&stored); err != nil {
		t.Fatalf("query published_at: %v", err)
	}
	if stored != "2025-01-03T09:00:00Z" {
		t.Fatalf("published_at = %q, want UTC RFC3339 value", stored)
	}
}

func TestPublishTimeParserPrefersArticlePublishedTimeAndFallsBackToDatePublished(t *testing.T) {
	preferred, err := parseArticlePublishedAt(`<html><head>` +
		`<script type="application/ld+json">{"datePublished":"2025-01-01T10:00:00+01:00"}</script>` +
		`<meta property="article:published_time" content="2025-01-03T10:00:00+01:00">` +
		`</head></html>`)
	if err != nil {
		t.Fatalf("parse preferred metadata: %v", err)
	}
	if got := formatPublishedAt(preferred); got != "2025-01-03T09:00:00Z" {
		t.Fatalf("preferred published_at = %q", got)
	}

	fallback, err := parseArticlePublishedAt(publishedJSONLDHTML("2025-01-02T10:00:00+01:00"))
	if err != nil {
		t.Fatalf("parse JSON-LD fallback: %v", err)
	}
	if got := formatPublishedAt(fallback); got != "2025-01-02T09:00:00Z" {
		t.Fatalf("fallback published_at = %q", got)
	}
}

func TestInitSchemaAddsAndBackfillsPublishedAtForCompatibleExistingDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "existing.db")
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer repo.Close()

	if _, err := repo.db.Exec(`
		CREATE TABLE article_sources (
			article_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL UNIQUE,
			discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE article_fetches (
			article_fetch_id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_source_id INTEGER NOT NULL UNIQUE REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
			html TEXT NOT NULL,
			fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO article_sources (url) VALUES ('https://example.com/backfill');
		INSERT INTO article_fetches (article_source_id, html) VALUES (1, ?);
	`, publishedMetaHTML("2025-01-04T10:00:00+01:00")); err != nil {
		t.Fatalf("seed compatible existing schema: %v", err)
	}

	if err := repo.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	assertColumnExists(t, repo, "article_sources", "published_at")

	var stored string
	if err := repo.db.QueryRow(`SELECT published_at FROM article_sources WHERE article_source_id=1`).Scan(&stored); err != nil {
		t.Fatalf("query backfilled published_at: %v", err)
	}
	if stored != "2025-01-04T09:00:00Z" {
		t.Fatalf("backfilled published_at = %q", stored)
	}
}

func TestInitSchemaBackfillFailsOnMissingPublicationMetadata(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "existing-invalid.db")
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer repo.Close()

	if _, err := repo.db.Exec(`
		CREATE TABLE article_sources (
			article_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL UNIQUE,
			discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE article_fetches (
			article_fetch_id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_source_id INTEGER NOT NULL UNIQUE REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
			html TEXT NOT NULL,
			fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO article_sources (url) VALUES ('https://example.com/invalid');
		INSERT INTO article_fetches (article_source_id, html) VALUES (1, '<html>missing metadata</html>');
	`); err != nil {
		t.Fatalf("seed compatible existing schema: %v", err)
	}

	err = repo.InitSchema()
	if err == nil {
		t.Fatalf("expected InitSchema to fail")
	}
	if !strings.Contains(err.Error(), "backfilling article_sources.published_at") || !strings.Contains(err.Error(), "missing parseable article publish metadata") {
		t.Fatalf("unexpected backfill error: %v", err)
	}
}

func TestInitSchemaBackfillFailsOnInvalidPublicationMetadata(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "existing-invalid-timestamp.db")
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer repo.Close()

	if _, err := repo.db.Exec(`
		CREATE TABLE article_sources (
			article_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL UNIQUE,
			discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE article_fetches (
			article_fetch_id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_source_id INTEGER NOT NULL UNIQUE REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
			html TEXT NOT NULL,
			fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO article_sources (url) VALUES ('https://example.com/invalid-timestamp');
		INSERT INTO article_fetches (article_source_id, html) VALUES (1, '<html><head><meta property="article:published_time" content="not-a-date"></head></html>');
	`); err != nil {
		t.Fatalf("seed compatible existing schema: %v", err)
	}

	err = repo.InitSchema()
	if err == nil {
		t.Fatalf("expected InitSchema to fail")
	}
	if !strings.Contains(err.Error(), "parsing article:published_time") {
		t.Fatalf("unexpected backfill error: %v", err)
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
