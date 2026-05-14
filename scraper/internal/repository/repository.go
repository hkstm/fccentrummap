package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/hkstm/fccentrummap/internal/models"
	sqlite "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}


func init() {
	sqlite.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, _ string) error {
		_, err := conn.ExecContext(context.Background(), "PRAGMA foreign_keys = ON", nil)
		return err
	})
}

func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS article_sources (
		article_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS article_fetches (
		article_fetch_id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_source_id INTEGER NOT NULL UNIQUE
			REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
		html TEXT NOT NULL,
		fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS article_texts (
		article_text_id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_fetch_id INTEGER NOT NULL UNIQUE
			REFERENCES article_fetches(article_fetch_id) ON DELETE CASCADE,
		cleaned_text TEXT NOT NULL,
		extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS audio_sources (
		audio_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_fetch_id INTEGER NOT NULL UNIQUE
			REFERENCES article_fetches(article_fetch_id) ON DELETE CASCADE,
		youtube_url TEXT NOT NULL,
		audio_format TEXT NOT NULL,
		mime_type TEXT NOT NULL,
		audio_blob BLOB NOT NULL,
		byte_size INTEGER NOT NULL,
		acquired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS audio_transcriptions (
		transcription_id INTEGER PRIMARY KEY AUTOINCREMENT,
		audio_source_id INTEGER NOT NULL
			REFERENCES audio_sources(audio_source_id) ON DELETE CASCADE,
		provider TEXT NOT NULL,
		language TEXT NOT NULL,
		http_status INTEGER NOT NULL,
		response_json TEXT NOT NULL CHECK(json_valid(response_json)),
		response_byte_size INTEGER NOT NULL,
		error_message TEXT,
		transcribed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(audio_source_id, provider, language)
	);

	CREATE TABLE IF NOT EXISTS spot_mentions (
		spot_mention_id INTEGER PRIMARY KEY AUTOINCREMENT,
		transcription_id INTEGER NOT NULL
			REFERENCES audio_transcriptions(transcription_id) ON DELETE CASCADE,
		place TEXT NOT NULL,
		sentence_start_timestamp REAL,
		original_sentence_start_timestamp REAL,
		refined_sentence_start_timestamp REAL,
		extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(transcription_id, place)
	);

	CREATE TABLE IF NOT EXISTS spot_google_geocodes (
		spot_google_geocode_id INTEGER PRIMARY KEY AUTOINCREMENT,
		spot_mention_id INTEGER NOT NULL UNIQUE
			REFERENCES spot_mentions(spot_mention_id) ON DELETE CASCADE,
		google_place_id TEXT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		formatted_address TEXT,
		status TEXT NOT NULL,
		geocoded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS presenters (
		presenter_id INTEGER PRIMARY KEY AUTOINCREMENT,
		presenter_name TEXT NOT NULL UNIQUE,
		materialized_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS article_presenters (
		article_source_id INTEGER NOT NULL
			REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
		presenter_id INTEGER NOT NULL
			REFERENCES presenters(presenter_id) ON DELETE CASCADE,
		linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (article_source_id, presenter_id)
	);

	CREATE TABLE IF NOT EXISTS article_spots (
		article_source_id INTEGER NOT NULL
			REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
		spot_google_geocode_id INTEGER NOT NULL
			REFERENCES spot_google_geocodes(spot_google_geocode_id) ON DELETE CASCADE,
		linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (article_source_id, spot_google_geocode_id)
	);

	`
	if _, err := r.db.Exec(schema); err != nil {
		return fmt.Errorf("initializing schema: %w", err)
	}

	return nil
}

func (r *Repository) UpsertArticleFetch(articleSourceID int64, html string) (int64, error) {
	var articleFetchID int64
	err := r.db.QueryRow(
		`INSERT INTO article_fetches (article_source_id, html)
		 VALUES (?, ?)
		 ON CONFLICT(article_source_id) DO UPDATE SET
			html = excluded.html,
			fetched_at = CURRENT_TIMESTAMP
		 RETURNING article_fetch_id`,
		articleSourceID,
		html,
	).Scan(&articleFetchID)
	if err != nil {
		return 0, fmt.Errorf("upserting article_fetches article_source_id=%d: %w", articleSourceID, err)
	}
	return articleFetchID, nil
}

func (r *Repository) UpsertArticleText(articleFetchID int64, cleanedText string) (int64, error) {
	var articleTextID int64
	err := r.db.QueryRow(
		`INSERT INTO article_texts (article_fetch_id, cleaned_text)
		 VALUES (?, ?)
		 ON CONFLICT(article_fetch_id) DO UPDATE SET
			cleaned_text = excluded.cleaned_text,
			extracted_at = CURRENT_TIMESTAMP
		 RETURNING article_text_id`,
		articleFetchID,
		cleanedText,
	).Scan(&articleTextID)
	if err != nil {
		return 0, fmt.Errorf("upserting article_texts article_fetch_id=%d: %w", articleFetchID, err)
	}
	return articleTextID, nil
}

func (r *Repository) UpsertSpotMention(transcriptionID int64, place string, sentenceStart, originalSentenceStart, refinedSentenceStart *float64) (int64, error) {
	var spotMentionID int64
	err := r.db.QueryRow(
		`INSERT INTO spot_mentions (
			transcription_id,
			place,
			sentence_start_timestamp,
			original_sentence_start_timestamp,
			refined_sentence_start_timestamp
		 ) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(transcription_id, place) DO UPDATE SET
			sentence_start_timestamp = excluded.sentence_start_timestamp,
			original_sentence_start_timestamp = excluded.original_sentence_start_timestamp,
			refined_sentence_start_timestamp = excluded.refined_sentence_start_timestamp,
			extracted_at = CURRENT_TIMESTAMP
		 RETURNING spot_mention_id`,
		transcriptionID,
		place,
		sentenceStart,
		originalSentenceStart,
		refinedSentenceStart,
	).Scan(&spotMentionID)
	if err != nil {
		return 0, fmt.Errorf("upserting spot_mentions transcription_id=%d place=%s: %w", transcriptionID, place, err)
	}
	return spotMentionID, nil
}

func (r *Repository) UpsertPresenter(name string) (int64, error) {
	var presenterID int64
	err := r.db.QueryRow(
		`INSERT INTO presenters (presenter_name)
		 VALUES (?)
		 ON CONFLICT(presenter_name) DO UPDATE SET
			materialized_at = CURRENT_TIMESTAMP
		 RETURNING presenter_id`,
		name,
	).Scan(&presenterID)
	if err != nil {
		return 0, fmt.Errorf("upserting presenters presenter_name=%s: %w", name, err)
	}
	return presenterID, nil
}

func (r *Repository) LinkArticlePresenter(articleSourceID, presenterID int64) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO article_presenters (article_source_id, presenter_id)
		 VALUES (?, ?)`,
		articleSourceID,
		presenterID,
	)
	if err != nil {
		return fmt.Errorf("linking article_presenters article_source_id=%d presenter_id=%d: %w", articleSourceID, presenterID, err)
	}
	return nil
}

func (r *Repository) UpsertSpotGoogleGeocode(spotMentionID int64, googlePlaceID *string, latitude, longitude float64, formattedAddress *string, status string) (int64, error) {
	var spotGoogleGeocodeID int64
	err := r.db.QueryRow(
		`INSERT INTO spot_google_geocodes (
			spot_mention_id,
			google_place_id,
			latitude,
			longitude,
			formatted_address,
			status
		 ) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(spot_mention_id) DO UPDATE SET
			google_place_id = excluded.google_place_id,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			formatted_address = excluded.formatted_address,
			status = excluded.status,
			geocoded_at = CURRENT_TIMESTAMP
		 RETURNING spot_google_geocode_id`,
		spotMentionID,
		googlePlaceID,
		latitude,
		longitude,
		formattedAddress,
		status,
	).Scan(&spotGoogleGeocodeID)
	if err != nil {
		return 0, fmt.Errorf("upserting spot_google_geocodes spot_mention_id=%d: %w", spotMentionID, err)
	}
	return spotGoogleGeocodeID, nil
}

func (r *Repository) UpsertAudioSource(articleFetchID int64, youtubeURL, audioFormat, mimeType string, audioBlob []byte) (int64, error) {
	var audioSourceID int64
	err := r.db.QueryRow(
		`INSERT INTO audio_sources (article_fetch_id, youtube_url, audio_format, mime_type, audio_blob, byte_size)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(article_fetch_id) DO UPDATE SET
			youtube_url = excluded.youtube_url,
			audio_format = excluded.audio_format,
			mime_type = excluded.mime_type,
			audio_blob = excluded.audio_blob,
			byte_size = excluded.byte_size,
			acquired_at = CURRENT_TIMESTAMP
		 RETURNING audio_source_id`,
		articleFetchID,
		youtubeURL,
		audioFormat,
		mimeType,
		audioBlob,
		len(audioBlob),
	).Scan(&audioSourceID)
	if err != nil {
		return 0, fmt.Errorf("upserting audio_sources article_fetch_id=%d: %w", articleFetchID, err)
	}
	return audioSourceID, nil
}

func (r *Repository) GetLatestAudioSource() (*models.ArticleAudioSource, error) {
	var src models.ArticleAudioSource
	err := r.db.QueryRow(
		`SELECT audio_source_id, article_fetch_id, youtube_url, audio_format, mime_type, audio_blob, byte_size, acquired_at
		 FROM audio_sources
		 WHERE audio_blob IS NOT NULL AND length(audio_blob) > 0
		 ORDER BY audio_source_id DESC
		 LIMIT 1`,
	).Scan(
		&src.AudioSourceID,
		&src.ArticleRawID,
		&src.YouTubeURL,
		&src.AudioFormat,
		&src.MIMEType,
		&src.AudioBlob,
		&src.ByteSize,
		&src.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying latest audio_sources: %w", err)
	}
	return &src, nil
}

func (r *Repository) UpsertAudioTranscription(t models.ArticleAudioTranscription) (int64, error) {
	var transcriptionID int64
	err := r.db.QueryRow(
		`INSERT INTO audio_transcriptions
			(audio_source_id, provider, language, http_status, response_json, response_byte_size, error_message)
		 VALUES (?, ?, ?, ?, json(?), ?, ?)
		 ON CONFLICT(audio_source_id, provider, language) DO UPDATE SET
			http_status = excluded.http_status,
			response_json = excluded.response_json,
			response_byte_size = excluded.response_byte_size,
			error_message = excluded.error_message,
			transcribed_at = CURRENT_TIMESTAMP
		 RETURNING transcription_id`,
		t.AudioSourceID,
		t.Provider,
		t.Language,
		t.HTTPStatus,
		t.ResponseJSON,
		t.ResponseByteSize,
		t.ErrorMessage,
	).Scan(&transcriptionID)
	if err != nil {
		return 0, fmt.Errorf("upserting audio_transcriptions audio_source_id=%d provider=%s language=%s: %w", t.AudioSourceID, t.Provider, t.Language, err)
	}
	return transcriptionID, nil
}

func (r *Repository) GetLatestAudioTranscription() (*models.ArticleAudioTranscription, error) {
	var t models.ArticleAudioTranscription
	var errMsg sql.NullString
	err := r.db.QueryRow(
		`SELECT transcription_id, audio_source_id, provider, language, http_status, response_json, response_byte_size, error_message, transcribed_at
		 FROM audio_transcriptions
		 ORDER BY transcription_id DESC
		 LIMIT 1`,
	).Scan(&t.TranscriptionID, &t.AudioSourceID, &t.Provider, &t.Language, &t.HTTPStatus, &t.ResponseJSON, &t.ResponseByteSize, &errMsg, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying latest audio_transcriptions: %w", err)
	}
	if errMsg.Valid {
		t.ErrorMessage = &errMsg.String
	}
	return &t, nil
}

func (r *Repository) GetArticleContextByTranscriptionID(transcriptionID int64) (string, int64, string, error) {
	var articleURL string
	var articleSourceID int64
	var cleanedText string
	err := r.db.QueryRow(
		`SELECT s.url, s.article_source_id, t.cleaned_text
		 FROM audio_transcriptions tr
		 JOIN audio_sources au ON au.audio_source_id = tr.audio_source_id
		 JOIN article_fetches f ON f.article_fetch_id = au.article_fetch_id
		 JOIN article_sources s ON s.article_source_id = f.article_source_id
		 JOIN article_texts t ON t.article_fetch_id = f.article_fetch_id
		 WHERE tr.transcription_id = ?`,
		transcriptionID,
	).Scan(&articleURL, &articleSourceID, &cleanedText)
	if err != nil {
		return "", 0, "", fmt.Errorf("querying article context by transcription_id=%d: %w", transcriptionID, err)
	}
	return articleURL, articleSourceID, cleanedText, nil
}

type SpotMentionForGeocode struct {
	SpotMentionID   int64
	ArticleSourceID int64
	Place           string
}

func (r *Repository) ListSpotMentionsWithoutGeocode() ([]SpotMentionForGeocode, error) {
	rows, err := r.db.Query(
		`SELECT sm.spot_mention_id, f.article_source_id, sm.place
		 FROM spot_mentions sm
		 JOIN audio_transcriptions tr ON tr.transcription_id = sm.transcription_id
		 JOIN audio_sources au ON au.audio_source_id = tr.audio_source_id
		 JOIN article_fetches f ON f.article_fetch_id = au.article_fetch_id
		 LEFT JOIN spot_google_geocodes sg ON sg.spot_mention_id = sm.spot_mention_id
		 WHERE sg.spot_google_geocode_id IS NULL
		 ORDER BY sm.spot_mention_id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying spot_mentions without geocode: %w", err)
	}
	defer rows.Close()
	out := []SpotMentionForGeocode{}
	for rows.Next() {
		var r SpotMentionForGeocode
		if err := rows.Scan(&r.SpotMentionID, &r.ArticleSourceID, &r.Place); err != nil {
			return nil, fmt.Errorf("scanning spot mention without geocode: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating spot mention without geocode: %w", err)
	}
	return out, nil
}

func (r *Repository) LinkArticleSpot(articleSourceID, spotGoogleGeocodeID int64) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO article_spots (article_source_id, spot_google_geocode_id)
		 VALUES (?, ?)`,
		articleSourceID,
		spotGoogleGeocodeID,
	)
	if err != nil {
		return fmt.Errorf("linking article_spots article_source_id=%d spot_google_geocode_id=%d: %w", articleSourceID, spotGoogleGeocodeID, err)
	}
	return nil
}

func (r *Repository) UpsertSpotGoogleGeocodeAndLinkArticleSpot(spotMentionID int64, googlePlaceID *string, latitude, longitude float64, formattedAddress *string, status string, articleSourceID int64) (int64, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("begin geocode/article_spot tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var spotGoogleGeocodeID int64
	err = tx.QueryRow(
		`INSERT INTO spot_google_geocodes (
			spot_mention_id,
			google_place_id,
			latitude,
			longitude,
			formatted_address,
			status
		 ) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(spot_mention_id) DO UPDATE SET
			google_place_id = excluded.google_place_id,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			formatted_address = excluded.formatted_address,
			status = excluded.status,
			geocoded_at = CURRENT_TIMESTAMP
		 RETURNING spot_google_geocode_id`,
		spotMentionID,
		googlePlaceID,
		latitude,
		longitude,
		formattedAddress,
		status,
	).Scan(&spotGoogleGeocodeID)
	if err != nil {
		return 0, fmt.Errorf("upserting spot_google_geocodes spot_mention_id=%d: %w", spotMentionID, err)
	}

	if _, err := tx.Exec(
		`INSERT OR IGNORE INTO article_spots (article_source_id, spot_google_geocode_id)
		 VALUES (?, ?)`,
		articleSourceID,
		spotGoogleGeocodeID,
	); err != nil {
		return 0, fmt.Errorf("linking article_spots article_source_id=%d spot_google_geocode_id=%d: %w", articleSourceID, spotGoogleGeocodeID, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit geocode/article_spot tx: %w", err)
	}
	return spotGoogleGeocodeID, nil
}

func (r *Repository) UpsertArticleSource(url string) (int64, error) {
	var articleSourceID int64
	err := r.db.QueryRow(
		`INSERT INTO article_sources (url)
		 VALUES (?)
		 ON CONFLICT(url) DO UPDATE SET
			discovered_at = article_sources.discovered_at
		 RETURNING article_source_id`,
		url,
	).Scan(&articleSourceID)
	if err != nil {
		return 0, fmt.Errorf("upserting article_sources url=%s: %w", url, err)
	}
	return articleSourceID, nil
}

func (r *Repository) ListArticleSources() ([]models.ArticleSource, error) {
	rows, err := r.db.Query(`SELECT article_source_id, url, discovered_at FROM article_sources ORDER BY article_source_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying article_sources: %w", err)
	}
	defer rows.Close()

	var out []models.ArticleSource
	for rows.Next() {
		var row models.ArticleSource
		if err := rows.Scan(&row.ArticleSourceID, &row.URL, &row.DiscoveredAt); err != nil {
			return nil, fmt.Errorf("scanning article_sources row: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating article_sources rows: %w", err)
	}
	return out, nil
}

func (r *Repository) ListArticleSourcesWithoutFetch() ([]models.ArticleSource, error) {
	rows, err := r.db.Query(
		`SELECT s.article_source_id, s.url, s.discovered_at
		 FROM article_sources s
		 LEFT JOIN article_fetches f ON f.article_source_id = s.article_source_id
		 WHERE f.article_fetch_id IS NULL
		 ORDER BY s.article_source_id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying article_sources without fetch: %w", err)
	}
	defer rows.Close()

	var out []models.ArticleSource
	for rows.Next() {
		var row models.ArticleSource
		if err := rows.Scan(&row.ArticleSourceID, &row.URL, &row.DiscoveredAt); err != nil {
			return nil, fmt.Errorf("scanning article_sources without fetch row: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating article_sources without fetch rows: %w", err)
	}
	return out, nil
}

func (r *Repository) ListArticleFetches() ([]models.ArticleFetch, error) {
	rows, err := r.db.Query(`SELECT article_fetch_id, article_source_id, html, fetched_at FROM article_fetches ORDER BY article_fetch_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying article_fetches: %w", err)
	}
	defer rows.Close()

	var out []models.ArticleFetch
	for rows.Next() {
		var row models.ArticleFetch
		if err := rows.Scan(&row.ArticleFetchID, &row.ArticleSourceID, &row.HTML, &row.FetchedAt); err != nil {
			return nil, fmt.Errorf("scanning article_fetches row: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating article_fetches rows: %w", err)
	}
	return out, nil
}

func (r *Repository) ExportData() (*models.ExportData, error) {
	rows, err := r.db.Query(`
		SELECT
			COALESCE(sgg.google_place_id, ''),
			COALESCE(sm.place, ''),
			NULLIF(p.presenter_name, ''),
			sgg.latitude,
			sgg.longitude,
			COALESCE(aus.youtube_url, ''),
			COALESCE(s.url, ''),
			sm.refined_sentence_start_timestamp,
			sm.original_sentence_start_timestamp,
			sm.sentence_start_timestamp
		FROM article_spots asp
		JOIN spot_google_geocodes sgg ON sgg.spot_google_geocode_id = asp.spot_google_geocode_id
		JOIN spot_mentions sm ON sm.spot_mention_id = sgg.spot_mention_id
		JOIN audio_transcriptions tr ON tr.transcription_id = sm.transcription_id
		JOIN audio_sources aus ON aus.audio_source_id = tr.audio_source_id
		JOIN article_sources s ON s.article_source_id = asp.article_source_id
		LEFT JOIN article_presenters ap ON ap.article_source_id = asp.article_source_id
		LEFT JOIN presenters p ON p.presenter_id = ap.presenter_id
	`)
	if err != nil {
		return nil, fmt.Errorf("querying export data: %w", err)
	}
	defer rows.Close()

	data := &models.ExportData{
		Spots:      []models.ExportSpot{},
		Presenters: []models.ExportPresenter{},
	}
	presenterSet := make(map[string]struct{})

	for rows.Next() {
		var (
			spot            models.ExportSpot
			rawYouTubeLink  string
			refinedTS       sql.NullFloat64
			originalTS      sql.NullFloat64
			sentenceStartTS sql.NullFloat64
		)
		var presenterName sql.NullString
		if err := rows.Scan(&spot.PlaceID, &spot.SpotName, &presenterName, &spot.Latitude, &spot.Longitude, &rawYouTubeLink, &spot.ArticleURL, &refinedTS, &originalTS, &sentenceStartTS); err != nil {
			return nil, fmt.Errorf("scanning export row: %w", err)
		}
		if presenterName.Valid {
			spot.PresenterName = presenterName.String
		}
		spot.YouTubeLink = withYouTubeTimestamp(rawYouTubeLink, refinedTS, originalTS, sentenceStartTS)
		data.Spots = append(data.Spots, spot)

		if presenterName.Valid {
			if _, ok := presenterSet[presenterName.String]; !ok {
				presenterSet[presenterName.String] = struct{}{}
				data.Presenters = append(data.Presenters, models.ExportPresenter{PresenterName: presenterName.String})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating export rows: %w", err)
	}

	slices.SortFunc(data.Spots, func(a, b models.ExportSpot) int {
		if a.PlaceID < b.PlaceID {
			return -1
		}
		if a.PlaceID > b.PlaceID {
			return 1
		}
		if a.PresenterName < b.PresenterName {
			return -1
		}
		if a.PresenterName > b.PresenterName {
			return 1
		}
		if a.SpotName < b.SpotName {
			return -1
		}
		if a.SpotName > b.SpotName {
			return 1
		}
		if a.YouTubeLink < b.YouTubeLink {
			return -1
		}
		if a.YouTubeLink > b.YouTubeLink {
			return 1
		}
		return 0
	})
	slices.SortFunc(data.Presenters, func(a, b models.ExportPresenter) int {
		if a.PresenterName < b.PresenterName {
			return -1
		}
		if a.PresenterName > b.PresenterName {
			return 1
		}
		return 0
	})

	return data, nil
}

func withYouTubeTimestamp(raw string, refinedTS, originalTS, sentenceStartTS sql.NullFloat64) string {
	ts := pickTimestamp(refinedTS, originalTS, sentenceStartTS)
	if ts == nil || *ts < 0 {
		return raw
	}
	videoID := extractYouTubeVideoID(raw)
	if videoID == "" {
		return raw
	}
	return fmt.Sprintf("https://youtu.be/%s?t=%d", videoID, int(*ts))
}

func pickTimestamp(refinedTS, originalTS, sentenceStartTS sql.NullFloat64) *float64 {
	if refinedTS.Valid {
		return &refinedTS.Float64
	}
	if originalTS.Valid {
		return &originalTS.Float64
	}
	if sentenceStartTS.Valid {
		return &sentenceStartTS.Float64
	}
	return nil
}

func extractYouTubeVideoID(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Host)
	switch host {
	case "youtu.be", "www.youtu.be":
		return strings.TrimPrefix(u.Path, "/")
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		return strings.TrimSpace(u.Query().Get("v"))
	default:
		return ""
	}
}
