package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	htmlstd "html"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

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
		published_at TIMESTAMP,
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
	if err := r.ensureArticleSourcePublishedAtColumn(); err != nil {
		return err
	}
	if err := r.backfillArticleSourcePublishedAt(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) ensureArticleSourcePublishedAtColumn() error {
	rows, err := r.db.Query(`PRAGMA table_info(article_sources)`)
	if err != nil {
		return fmt.Errorf("inspecting article_sources schema: %w", err)
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
			return fmt.Errorf("scanning article_sources schema: %w", err)
		}
		if name == "published_at" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating article_sources schema: %w", err)
	}
	if _, err := r.db.Exec(`ALTER TABLE article_sources ADD COLUMN published_at TIMESTAMP`); err != nil {
		return fmt.Errorf("adding article_sources.published_at: %w", err)
	}
	return nil
}

func (r *Repository) backfillArticleSourcePublishedAt() error {
	rows, err := r.db.Query(`
		SELECT s.article_source_id, s.url, f.html
		FROM article_sources s
		JOIN article_fetches f ON f.article_source_id = s.article_source_id
		WHERE s.published_at IS NULL
		ORDER BY s.article_source_id ASC`)
	if err != nil {
		return fmt.Errorf("querying article_sources published_at backfill candidates: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		articleSourceID int64
		url             string
		html            string
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.articleSourceID, &c.url, &c.html); err != nil {
			return fmt.Errorf("scanning article_sources published_at backfill candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating article_sources published_at backfill candidates: %w", err)
	}

	for _, c := range candidates {
		publishedAt, err := parseArticlePublishedAt(c.html)
		if err != nil {
			return fmt.Errorf("backfilling article_sources.published_at article_source_id=%d url=%s: %w", c.articleSourceID, c.url, err)
		}
		if err := r.setArticleSourcePublishedAt(c.articleSourceID, publishedAt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) setArticleSourcePublishedAt(articleSourceID int64, publishedAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE article_sources SET published_at = ? WHERE article_source_id = ?`,
		formatPublishedAt(publishedAt),
		articleSourceID,
	)
	if err != nil {
		return fmt.Errorf("updating article_sources.published_at article_source_id=%d: %w", articleSourceID, err)
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
	if publishedAt, err := parseArticlePublishedAt(html); err == nil {
		if err := r.setArticleSourcePublishedAt(articleSourceID, publishedAt); err != nil {
			return 0, err
		}
	}
	return articleFetchID, nil
}

var (
	metaTagRe           = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	attrRe              = regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("[^"]*"|'[^']*')`)
	jsonLDScriptRe      = regexp.MustCompile(`(?is)<script\b[^>]*type\s*=\s*(?:"application/ld\+json"|'application/ld\+json')[^>]*>(.*?)</script>`)
	datePublishedRe     = regexp.MustCompile(`(?is)"datePublished"\s*:\s*"([^"]+)"`)
	sqliteTimestampForm = []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
)

func parseArticlePublishedAt(htmlBody string) (time.Time, error) {
	if published, ok, err := parseArticlePublishedAtMeta(htmlBody); ok || err != nil {
		return published, err
	}
	if published, ok, err := parseArticlePublishedAtJSONLD(htmlBody); ok || err != nil {
		return published, err
	}
	return time.Time{}, errors.New("missing parseable article publish metadata")
}

func parseArticlePublishedAtMeta(htmlBody string) (time.Time, bool, error) {
	for _, tag := range metaTagRe.FindAllString(htmlBody, -1) {
		attrs := parseHTMLAttrs(tag)
		property := strings.TrimSpace(attrs["property"])
		name := strings.TrimSpace(attrs["name"])
		if property != "article:published_time" && name != "article:published_time" {
			continue
		}
		raw := strings.TrimSpace(attrs["content"])
		if raw == "" {
			return time.Time{}, true, errors.New("article:published_time metadata is missing content")
		}
		published, err := parsePublishedTimestamp(raw)
		if err != nil {
			return time.Time{}, true, fmt.Errorf("parsing article:published_time %q: %w", raw, err)
		}
		return published, true, nil
	}
	return time.Time{}, false, nil
}

func parseArticlePublishedAtJSONLD(htmlBody string) (time.Time, bool, error) {
	for _, match := range jsonLDScriptRe.FindAllStringSubmatch(htmlBody, -1) {
		if len(match) < 2 {
			continue
		}
		body := htmlstd.UnescapeString(strings.TrimSpace(match[1]))
		var doc any
		if err := json.Unmarshal([]byte(body), &doc); err == nil {
			if raw, ok := findDatePublished(doc); ok {
				published, err := parsePublishedTimestamp(raw)
				if err != nil {
					return time.Time{}, true, fmt.Errorf("parsing datePublished %q: %w", raw, err)
				}
				return published, true, nil
			}
		}
		if match := datePublishedRe.FindStringSubmatch(body); len(match) == 2 {
			published, err := parsePublishedTimestamp(htmlstd.UnescapeString(match[1]))
			if err != nil {
				return time.Time{}, true, fmt.Errorf("parsing datePublished %q: %w", match[1], err)
			}
			return published, true, nil
		}
	}
	return time.Time{}, false, nil
}

func findDatePublished(v any) (string, bool) {
	switch t := v.(type) {
	case map[string]any:
		if raw, ok := t["datePublished"].(string); ok && strings.TrimSpace(raw) != "" {
			return strings.TrimSpace(raw), true
		}
		for _, child := range t {
			if raw, ok := findDatePublished(child); ok {
				return raw, true
			}
		}
	case []any:
		for _, child := range t {
			if raw, ok := findDatePublished(child); ok {
				return raw, true
			}
		}
	}
	return "", false
}

func parseHTMLAttrs(tag string) map[string]string {
	attrs := map[string]string{}
	for _, match := range attrRe.FindAllStringSubmatch(tag, -1) {
		if len(match) != 3 {
			continue
		}
		key := strings.ToLower(match[1])
		value := match[2]
		if len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		attrs[key] = htmlstd.UnescapeString(value)
	}
	return attrs
}

func parsePublishedTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range sqliteTimestampForm {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format")
}

func parseStoredPublishedAt(raw string) (time.Time, error) {
	return parsePublishedTimestamp(raw)
}

func formatPublishedAt(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
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

func (r *Repository) ListAudioSourcesPendingTranscription() ([]models.ArticleAudioSource, error) {
	rows, err := r.db.Query(
		`SELECT s.audio_source_id, s.article_fetch_id, s.youtube_url, s.audio_format, s.mime_type, s.audio_blob, s.byte_size, s.acquired_at
		 FROM audio_sources s
		 LEFT JOIN audio_transcriptions t ON t.audio_source_id = s.audio_source_id
		 WHERE s.audio_blob IS NOT NULL AND length(s.audio_blob) > 0
		   AND t.transcription_id IS NULL
		 ORDER BY s.audio_source_id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending audio_sources: %w", err)
	}
	defer rows.Close()
	var out []models.ArticleAudioSource

	for rows.Next() {
		var src models.ArticleAudioSource
		if err := rows.Scan(
			&src.AudioSourceID,
			&src.ArticleRawID,
			&src.YouTubeURL,
			&src.AudioFormat,
			&src.MIMEType,
			&src.AudioBlob,
			&src.ByteSize,
			&src.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning pending article audio sources: %w", err)
		}
		out = append(out, src)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pending article audio sources: %w", err)
	}
	return out, nil
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

func (r *Repository) ListTranscriptionsPendingExtraction() ([]models.ArticleAudioTranscription, error) {
	rows, err := r.db.Query(
		`SELECT tr.transcription_id, tr.audio_source_id, tr.provider, tr.language, tr.http_status, tr.response_json, tr.response_byte_size, tr.error_message, tr.transcribed_at
		 FROM audio_transcriptions tr
		 JOIN audio_sources au ON au.audio_source_id = tr.audio_source_id
		 JOIN article_fetches f ON f.article_fetch_id = au.article_fetch_id
		 WHERE NOT EXISTS (
			 SELECT 1 FROM spot_mentions sm WHERE sm.transcription_id = tr.transcription_id
		 ) AND NOT EXISTS (
			 SELECT 1 FROM article_presenters ap WHERE ap.article_source_id = f.article_source_id
		 )
		 ORDER BY tr.transcription_id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending transcriptions: %w", err)
	}
	defer rows.Close()

	var results []models.ArticleAudioTranscription
	for rows.Next() {
		var t models.ArticleAudioTranscription
		var errMsg sql.NullString
		err := rows.Scan(&t.TranscriptionID, &t.AudioSourceID, &t.Provider, &t.Language, &t.HTTPStatus, &t.ResponseJSON, &t.ResponseByteSize, &errMsg, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning pending transcription: %w", err)
		}
		if errMsg.Valid {
			t.ErrorMessage = &errMsg.String
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pending transcriptions: %w", err)
	}
	return results, nil
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
			s.published_at,
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
	presenterLatestPublishedAt := make(map[string]time.Time)

	for rows.Next() {
		var (
			spot             models.ExportSpot
			rawYouTubeLink   string
			publishedAtValue sql.NullString
			refinedTS        sql.NullFloat64
			originalTS       sql.NullFloat64
			sentenceStartTS  sql.NullFloat64
		)
		var presenterName sql.NullString
		if err := rows.Scan(&spot.PlaceID, &spot.SpotName, &presenterName, &spot.Latitude, &spot.Longitude, &rawYouTubeLink, &spot.ArticleURL, &publishedAtValue, &refinedTS, &originalTS, &sentenceStartTS); err != nil {
			return nil, fmt.Errorf("scanning export row: %w", err)
		}
		if !publishedAtValue.Valid || strings.TrimSpace(publishedAtValue.String) == "" {
			return nil, fmt.Errorf("exportable article url=%s has no stored publication time", spot.ArticleURL)
		}
		publishedAt, err := parseStoredPublishedAt(publishedAtValue.String)
		if err != nil {
			return nil, fmt.Errorf("exportable article url=%s has unparseable stored publication time %q: %w", spot.ArticleURL, publishedAtValue.String, err)
		}
		if presenterName.Valid {
			spot.PresenterName = presenterName.String
		}
		spot.YouTubeLink = withYouTubeTimestamp(rawYouTubeLink, refinedTS, originalTS, sentenceStartTS)
		data.Spots = append(data.Spots, spot)

		if presenterName.Valid {
			current, ok := presenterLatestPublishedAt[presenterName.String]
			if !ok || publishedAt.After(current) {
				presenterLatestPublishedAt[presenterName.String] = publishedAt
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
	for presenterName := range presenterLatestPublishedAt {
		data.Presenters = append(data.Presenters, models.ExportPresenter{PresenterName: presenterName})
	}
	slices.SortFunc(data.Presenters, func(a, b models.ExportPresenter) int {
		aPublished := presenterLatestPublishedAt[a.PresenterName]
		bPublished := presenterLatestPublishedAt[b.PresenterName]
		if aPublished.After(bPublished) {
			return -1
		}
		if aPublished.Before(bPublished) {
			return 1
		}
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
