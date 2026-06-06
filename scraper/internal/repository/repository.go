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
	"github.com/hkstm/fccentrummap/internal/youtube"
	sqlite "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

const geminiDirectMinimumGeocodeConfidence = 0.9

type GeminiDirectInput struct {
	ArticleSourceID int64
	ArticleURL      string
	YouTubeURL      string
}

type PlaceTypeMetadata struct {
	PrimaryType            *string
	PrimaryTypeDisplayName *string
}

type GeminiDirectSpotMentionInput struct {
	ArticleSourceID         int64
	ArticleURL              string
	YouTubeURL              string
	Place                   string
	Address                 *string
	YouTubeTimestampSeconds *float64
	Evidence                *string
	Confidence              *float64
	Model                   *string
}

type GeminiDirectExtractionSpot struct {
	Place                   string
	Address                 *string
	YouTubeTimestampSeconds *float64
	Evidence                string
	Confidence              *float64
	Model                   string
}

type GeminiDirectExtractionArticle struct {
	ArticleSourceID int64
	ArticleURL      string
	YouTubeURL      string
	PresenterName   *string
	Model           string
	Spots           []GeminiDirectExtractionSpot
}

type SpotCategoryMapping struct {
	PrimaryTypeDisplayName string
	CategoryName           string
	Confidence             *float64
	Reason                 *string
	Model                  string
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

	CREATE TABLE IF NOT EXISTS gemini_direct_spot_mentions (
		gemini_direct_spot_mention_id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_source_id INTEGER NOT NULL
			REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
		youtube_url TEXT NOT NULL,
		place TEXT NOT NULL,
		address TEXT,
		youtube_timestamp_seconds REAL CHECK (youtube_timestamp_seconds IS NULL OR youtube_timestamp_seconds >= 0),
		evidence TEXT,
		confidence REAL CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1)),
		model TEXT,
		extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(article_source_id, place)
	);

	CREATE TABLE IF NOT EXISTS gemini_direct_spot_google_geocodes (
		gemini_direct_spot_google_geocode_id INTEGER PRIMARY KEY AUTOINCREMENT,
		gemini_direct_spot_mention_id INTEGER NOT NULL UNIQUE
			REFERENCES gemini_direct_spot_mentions(gemini_direct_spot_mention_id) ON DELETE CASCADE,
		google_place_id TEXT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		formatted_address TEXT,
		primary_type TEXT,
		primary_type_display_name TEXT,
		status TEXT NOT NULL,
		geocoded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS gemini_direct_article_spots (
		article_source_id INTEGER NOT NULL
			REFERENCES article_sources(article_source_id) ON DELETE CASCADE,
		gemini_direct_spot_google_geocode_id INTEGER NOT NULL
			REFERENCES gemini_direct_spot_google_geocodes(gemini_direct_spot_google_geocode_id) ON DELETE CASCADE,
		linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (article_source_id, gemini_direct_spot_google_geocode_id)
	);

	CREATE TABLE IF NOT EXISTS gemini_direct_spot_corrections (
		spot_id TEXT PRIMARY KEY CHECK (trim(spot_id) <> ''),
		spot_name TEXT,
		place_id TEXT,
		latitude REAL,
		longitude REAL,
		youtube_timestamp_seconds INTEGER CHECK (youtube_timestamp_seconds IS NULL OR youtube_timestamp_seconds >= 0),
		hidden INTEGER NOT NULL DEFAULT 0 CHECK (hidden IN (0, 1)),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (
			((place_id IS NULL OR trim(place_id) = '') AND latitude IS NULL AND longitude IS NULL)
			OR (place_id IS NOT NULL AND trim(place_id) <> '' AND latitude IS NOT NULL AND longitude IS NOT NULL)
		)
	);

	CREATE TABLE IF NOT EXISTS spot_category_mappings (
		primary_type_display_name TEXT PRIMARY KEY CHECK (trim(primary_type_display_name) <> ''),
		category_name TEXT NOT NULL CHECK (trim(category_name) <> ''),
		confidence REAL CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1)),
		reason TEXT,
		model TEXT,
		mapped_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	`
	if _, err := r.db.Exec(schema); err != nil {
		return fmt.Errorf("initializing schema: %w", err)
	}
	if err := r.ensureArticleSourcePublishedAtColumn(); err != nil {
		return err
	}
	if err := r.ensureGeminiDirectAddressColumn(); err != nil {
		return err
	}
	if err := r.ensureGeminiDirectGeocodeTypeColumns(); err != nil {
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

func (r *Repository) ensureGeminiDirectAddressColumn() error {
	exists, err := r.columnExists("gemini_direct_spot_mentions", "address")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := r.db.Exec(`ALTER TABLE gemini_direct_spot_mentions ADD COLUMN address TEXT`); err != nil {
		return fmt.Errorf("adding gemini_direct_spot_mentions.address: %w", err)
	}
	return nil
}

func (r *Repository) ensureGeminiDirectGeocodeTypeColumns() error {
	for _, col := range []struct {
		name string
		ddl  string
	}{
		{name: "primary_type", ddl: `ALTER TABLE gemini_direct_spot_google_geocodes ADD COLUMN primary_type TEXT`},
		{name: "primary_type_display_name", ddl: `ALTER TABLE gemini_direct_spot_google_geocodes ADD COLUMN primary_type_display_name TEXT`},
	} {
		exists, err := r.columnExists("gemini_direct_spot_google_geocodes", col.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := r.db.Exec(col.ddl); err != nil {
			return fmt.Errorf("adding gemini_direct_spot_google_geocodes.%s: %w", col.name, err)
		}
	}
	return nil
}

func (r *Repository) columnExists(table, column string) (bool, error) {
	rows, err := r.db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, fmt.Errorf("inspecting %s schema: %w", table, err)
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
			return false, fmt.Errorf("scanning %s schema: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterating %s schema: %w", table, err)
	}
	return false, nil
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

func (r *Repository) UpsertGeminiDirectSpotMention(input GeminiDirectSpotMentionInput) (int64, error) {
	place := strings.TrimSpace(input.Place)
	if place == "" {
		return 0, fmt.Errorf("Gemini-direct spot mention place is required")
	}
	youtubeURL := strings.TrimSpace(input.YouTubeURL)
	if youtubeURL == "" {
		return 0, fmt.Errorf("Gemini-direct spot mention YouTube URL is required")
	}
	if input.ArticleSourceID <= 0 {
		return 0, fmt.Errorf("Gemini-direct spot mention article_source_id is required")
	}
	var addressPtr *string
	if input.Address != nil {
		address := strings.TrimSpace(*input.Address)
		if address == "" {
			return 0, fmt.Errorf("Gemini-direct spot mention address must be non-empty when provided")
		}
		addressPtr = &address
	}
	if input.YouTubeTimestampSeconds != nil && *input.YouTubeTimestampSeconds < 0 {
		return 0, fmt.Errorf("Gemini-direct spot mention youtube_timestamp_seconds must be >= 0")
	}
	if input.Confidence != nil && (*input.Confidence < 0 || *input.Confidence > 1) {
		return 0, fmt.Errorf("Gemini-direct spot mention confidence must be between 0 and 1")
	}
	if strings.TrimSpace(input.ArticleURL) != "" {
		var storedURL string
		if err := r.db.QueryRow(`SELECT url FROM article_sources WHERE article_source_id = ?`, input.ArticleSourceID).Scan(&storedURL); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("Gemini-direct article_source_id=%d not found", input.ArticleSourceID)
			}
			return 0, fmt.Errorf("querying article_source_id=%d: %w", input.ArticleSourceID, err)
		}
		if strings.TrimSpace(storedURL) != strings.TrimSpace(input.ArticleURL) {
			return 0, fmt.Errorf("Gemini-direct article URL mismatch for article_source_id=%d", input.ArticleSourceID)
		}
	}
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO gemini_direct_spot_mentions (
			article_source_id,
			youtube_url,
			place,
			address,
			youtube_timestamp_seconds,
			evidence,
			confidence,
			model
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(article_source_id, place) DO UPDATE SET
			youtube_url = excluded.youtube_url,
			address = excluded.address,
			youtube_timestamp_seconds = excluded.youtube_timestamp_seconds,
			evidence = excluded.evidence,
			confidence = excluded.confidence,
			model = excluded.model,
			extracted_at = CURRENT_TIMESTAMP
		 RETURNING gemini_direct_spot_mention_id`,
		input.ArticleSourceID,
		youtubeURL,
		place,
		addressPtr,
		input.YouTubeTimestampSeconds,
		input.Evidence,
		input.Confidence,
		input.Model,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upserting gemini_direct_spot_mentions article_source_id=%d place=%s: %w", input.ArticleSourceID, place, err)
	}
	return id, nil
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

type SpotMentionForGeocode struct {
	SpotMentionID           int64
	ArticleSourceID         int64
	Place                   string
	Address                 *string
	YouTubeURL              string
	YouTubeTimestampSeconds *float64
}

func (r *Repository) ListSpotMentionsWithoutGeocodeForSource(source models.SpotSource) ([]SpotMentionForGeocode, error) {
	if _, err := models.NormalizeSpotSource(string(source)); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(
		`SELECT gsm.gemini_direct_spot_mention_id, gsm.article_source_id, gsm.place, gsm.address, gsm.youtube_url, gsm.youtube_timestamp_seconds
		 FROM gemini_direct_spot_mentions gsm
		 LEFT JOIN gemini_direct_spot_google_geocodes gsg ON gsg.gemini_direct_spot_mention_id = gsm.gemini_direct_spot_mention_id
		 WHERE gsg.gemini_direct_spot_google_geocode_id IS NULL
		 ORDER BY gsm.gemini_direct_spot_mention_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying Gemini-direct spot mentions without geocode: %w", err)
	}
	defer rows.Close()
	out := []SpotMentionForGeocode{}
	for rows.Next() {
		var row SpotMentionForGeocode
		var ts sql.NullFloat64
		var address sql.NullString
		if err := rows.Scan(&row.SpotMentionID, &row.ArticleSourceID, &row.Place, &address, &row.YouTubeURL, &ts); err != nil {
			return nil, fmt.Errorf("scanning Gemini-direct spot mention without geocode: %w", err)
		}
		if ts.Valid {
			row.YouTubeTimestampSeconds = &ts.Float64
		}
		if address.Valid && strings.TrimSpace(address.String) != "" {
			addr := strings.TrimSpace(address.String)
			row.Address = &addr
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating Gemini-direct spot mentions without geocode: %w", err)
	}
	return out, nil
}

func (r *Repository) UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(source models.SpotSource, mentionID int64, googlePlaceID *string, latitude, longitude float64, formattedAddress *string, status string, articleSourceID int64, metadata PlaceTypeMetadata) (int64, error) {
	if _, err := models.NormalizeSpotSource(string(source)); err != nil {
		return 0, err
	}
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("begin Gemini-direct geocode/article_spot tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var geocodeID int64
	err = tx.QueryRow(
		`INSERT INTO gemini_direct_spot_google_geocodes (
			gemini_direct_spot_mention_id,
			google_place_id,
			latitude,
			longitude,
			formatted_address,
			primary_type,
			primary_type_display_name,
			status
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(gemini_direct_spot_mention_id) DO UPDATE SET
			google_place_id = excluded.google_place_id,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			formatted_address = excluded.formatted_address,
			primary_type = excluded.primary_type,
			primary_type_display_name = excluded.primary_type_display_name,
			status = excluded.status,
			geocoded_at = CURRENT_TIMESTAMP
		 RETURNING gemini_direct_spot_google_geocode_id`,
		mentionID,
		googlePlaceID,
		latitude,
		longitude,
		formattedAddress,
		metadata.PrimaryType,
		metadata.PrimaryTypeDisplayName,
		status,
	).Scan(&geocodeID)
	if err != nil {
		return 0, fmt.Errorf("upserting gemini_direct_spot_google_geocodes gemini_direct_spot_mention_id=%d: %w", mentionID, err)
	}

	if _, err := tx.Exec(
		`INSERT OR IGNORE INTO gemini_direct_article_spots (article_source_id, gemini_direct_spot_google_geocode_id)
		 VALUES (?, ?)`,
		articleSourceID,
		geocodeID,
	); err != nil {
		return 0, fmt.Errorf("linking gemini_direct_article_spots article_source_id=%d geocode_id=%d: %w", articleSourceID, geocodeID, err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit Gemini-direct geocode/article_spot tx: %w", err)
	}
	return geocodeID, nil
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

func (r *Repository) ListGeminiDirectInputs() ([]GeminiDirectInput, error) {
	rows, err := r.db.Query(`
		SELECT s.article_source_id, COALESCE(s.url, ''), COALESCE(f.html, '')
		FROM article_sources s
		LEFT JOIN article_fetches f ON f.article_source_id = s.article_source_id
		ORDER BY s.article_source_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying Gemini-direct inputs: %w", err)
	}
	defer rows.Close()

	var out []GeminiDirectInput
	for rows.Next() {
		var row GeminiDirectInput
		var html string
		if err := rows.Scan(&row.ArticleSourceID, &row.ArticleURL, &html); err != nil {
			return nil, fmt.Errorf("scanning Gemini-direct input: %w", err)
		}
		if strings.TrimSpace(row.YouTubeURL) == "" {
			if videoID, ok := youtube.ExtractVideoID(html); ok {
				row.YouTubeURL = youtube.WatchURL(videoID)
			}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating Gemini-direct inputs: %w", err)
	}
	return out, nil
}

func StableSpotID(articleSourceID, spotMentionID int64) string {
	return fmt.Sprintf("%d:%d", articleSourceID, spotMentionID)
}

func (r *Repository) GetSpotCorrectionForSource(source models.SpotSource, spotID string) (*models.SpotCorrection, error) {
	if _, err := models.NormalizeSpotSource(string(source)); err != nil {
		return nil, err
	}
	return r.getSpotCorrectionFromTable("gemini_direct_spot_corrections", spotID)
}

func (r *Repository) getSpotCorrectionFromTable(table, spotID string) (*models.SpotCorrection, error) {
	var c models.SpotCorrection
	var spotName, placeID sql.NullString
	var lat, lng sql.NullFloat64
	var ts sql.NullInt64
	var hidden sql.NullBool
	err := r.db.QueryRow(
		fmt.Sprintf(`SELECT spot_id, spot_name, place_id, latitude, longitude, youtube_timestamp_seconds, hidden, created_at, updated_at
		 FROM %s
		 WHERE spot_id = ?`, table),
		strings.TrimSpace(spotID),
	).Scan(&c.SpotID, &spotName, &placeID, &lat, &lng, &ts, &hidden, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying %s spot_id=%s: %w", table, spotID, err)
	}
	if spotName.Valid {
		c.SpotName = &spotName.String
	}
	if placeID.Valid {
		c.PlaceID = &placeID.String
	}
	if lat.Valid {
		c.Latitude = &lat.Float64
	}
	if lng.Valid {
		c.Longitude = &lng.Float64
	}
	if ts.Valid {
		c.YouTubeTimestampSeconds = &ts.Int64
	}
	if hidden.Valid {
		c.Hidden = hidden.Bool
	}
	return &c, nil
}

func (r *Repository) UpsertSpotCorrectionForSource(source models.SpotSource, c models.SpotCorrection) error {
	if _, err := models.NormalizeSpotSource(string(source)); err != nil {
		return err
	}
	return r.upsertSpotCorrectionIntoTable("gemini_direct_spot_corrections", c)
}

func (r *Repository) upsertSpotCorrectionIntoTable(table string, c models.SpotCorrection) error {
	spotID := strings.TrimSpace(c.SpotID)
	if spotID == "" {
		return fmt.Errorf("spot correction spot_id is required")
	}
	if c.YouTubeTimestampSeconds != nil && *c.YouTubeTimestampSeconds < 0 {
		return fmt.Errorf("spot correction youtube_timestamp_seconds must be >= 0")
	}
	if (c.Latitude == nil) != (c.Longitude == nil) {
		return fmt.Errorf("spot correction latitude and longitude must be provided together")
	}
	hasPlace := c.PlaceID != nil && strings.TrimSpace(*c.PlaceID) != ""
	hasCoords := c.Latitude != nil && c.Longitude != nil
	if hasPlace != hasCoords {
		return fmt.Errorf("spot correction place_id and coordinates must be provided together")
	}
	_, err := r.db.Exec(
		fmt.Sprintf(`INSERT INTO %s (spot_id, spot_name, place_id, latitude, longitude, youtube_timestamp_seconds, hidden)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(spot_id) DO UPDATE SET
			spot_name = excluded.spot_name,
			place_id = excluded.place_id,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			youtube_timestamp_seconds = excluded.youtube_timestamp_seconds,
			hidden = excluded.hidden,
			updated_at = CURRENT_TIMESTAMP`, table),
		spotID,
		c.SpotName,
		c.PlaceID,
		c.Latitude,
		c.Longitude,
		c.YouTubeTimestampSeconds,
		c.Hidden,
	)
	if err != nil {
		return fmt.Errorf("upserting %s spot_id=%s: %w", table, spotID, err)
	}
	return nil
}

func (r *Repository) GetSpotCorrectionTargetForSource(source models.SpotSource, spotID string) (*models.SpotCorrectionTarget, error) {
	if _, err := models.NormalizeSpotSource(string(source)); err != nil {
		return nil, err
	}
	return r.getGeminiDirectSpotCorrectionTarget(spotID)
}

func (r *Repository) getGeminiDirectSpotCorrectionTarget(spotID string) (*models.SpotCorrectionTarget, error) {
	row := r.db.QueryRow(`
		SELECT
			CAST(gasp.article_source_id AS TEXT) || ':' || CAST(gsm.gemini_direct_spot_mention_id AS TEXT),
			COALESCE(gsm.place, ''),
			COALESCE(gsg.google_place_id, ''),
			gsg.latitude,
			gsg.longitude,
			COALESCE(gsm.youtube_url, ''),
			COALESCE(s.url, ''),
			NULLIF(p.presenter_name, ''),
			gsm.youtube_timestamp_seconds,
			gc.spot_name,
			gc.place_id,
			gc.latitude,
			gc.longitude,
			gc.youtube_timestamp_seconds,
			gc.hidden,
			gc.created_at,
			gc.updated_at
		FROM gemini_direct_article_spots gasp
		JOIN gemini_direct_spot_google_geocodes gsg ON gsg.gemini_direct_spot_google_geocode_id = gasp.gemini_direct_spot_google_geocode_id
		JOIN gemini_direct_spot_mentions gsm ON gsm.gemini_direct_spot_mention_id = gsg.gemini_direct_spot_mention_id
		JOIN article_sources s ON s.article_source_id = gasp.article_source_id
		LEFT JOIN article_presenters ap ON ap.article_source_id = gasp.article_source_id
		LEFT JOIN presenters p ON p.presenter_id = ap.presenter_id
		LEFT JOIN gemini_direct_spot_corrections gc ON gc.spot_id = CAST(gasp.article_source_id AS TEXT) || ':' || CAST(gsm.gemini_direct_spot_mention_id AS TEXT)
		WHERE CAST(gasp.article_source_id AS TEXT) || ':' || CAST(gsm.gemini_direct_spot_mention_id AS TEXT) = ?`,
		strings.TrimSpace(spotID),
	)

	var target models.SpotCorrectionTarget
	var presenterName sql.NullString
	var sourceTS sql.NullFloat64
	var cSpotName, cPlaceID sql.NullString
	var cLat, cLng sql.NullFloat64
	var cTimestamp sql.NullInt64
	var cHidden sql.NullBool
	var cCreated, cUpdated sql.NullTime
	if err := row.Scan(
		&target.SpotID,
		&target.SourceSpotName,
		&target.SourcePlaceID,
		&target.SourceLatitude,
		&target.SourceLongitude,
		&target.SourceYouTubeLink,
		&target.ArticleURL,
		&presenterName,
		&sourceTS,
		&cSpotName,
		&cPlaceID,
		&cLat,
		&cLng,
		&cTimestamp,
		&cHidden,
		&cCreated,
		&cUpdated,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying Gemini-direct spot correction target spot_id=%s: %w", spotID, err)
	}
	if presenterName.Valid {
		target.PresenterName = presenterName.String
	}
	target.SourceYouTubeTimestampSecs = pickedTimestampSeconds(sourceTS, sql.NullFloat64{}, sql.NullFloat64{})
	target.EffectiveSpotName = target.SourceSpotName
	target.EffectivePlaceID = target.SourcePlaceID
	target.EffectiveLatitude = target.SourceLatitude
	target.EffectiveLongitude = target.SourceLongitude
	target.EffectiveYouTubeTimestampSecs = target.SourceYouTubeTimestampSecs
	target.EffectiveYouTubeLink = withYouTubeTimestamp(target.SourceYouTubeLink, sourceTS, sql.NullFloat64{}, sql.NullFloat64{})
	applyCorrectionToTarget(&target, cSpotName, cPlaceID, cLat, cLng, cTimestamp, cHidden, cCreated, cUpdated)
	return &target, nil
}

func applyCorrectionToTarget(target *models.SpotCorrectionTarget, cSpotName, cPlaceID sql.NullString, cLat, cLng sql.NullFloat64, cTimestamp sql.NullInt64, cHidden sql.NullBool, cCreated, cUpdated sql.NullTime) {
	if !(cSpotName.Valid || cPlaceID.Valid || cLat.Valid || cLng.Valid || cTimestamp.Valid || cHidden.Valid) {
		return
	}
	correction := &models.SpotCorrection{SpotID: target.SpotID}
	if cSpotName.Valid {
		correction.SpotName = &cSpotName.String
		target.EffectiveSpotName = cSpotName.String
	}
	if cPlaceID.Valid {
		correction.PlaceID = &cPlaceID.String
		if strings.TrimSpace(cPlaceID.String) != "" {
			target.EffectivePlaceID = cPlaceID.String
		}
	}
	if cLat.Valid {
		correction.Latitude = &cLat.Float64
		target.EffectiveLatitude = cLat.Float64
	}
	if cLng.Valid {
		correction.Longitude = &cLng.Float64
		target.EffectiveLongitude = cLng.Float64
	}
	if cTimestamp.Valid {
		correction.YouTubeTimestampSeconds = &cTimestamp.Int64
		target.EffectiveYouTubeTimestampSecs = &cTimestamp.Int64
		if cTimestamp.Int64 >= 0 {
			target.EffectiveYouTubeLink = withYouTubeTimestampSeconds(target.SourceYouTubeLink, cTimestamp.Int64)
		}
	}
	if cHidden.Valid {
		correction.Hidden = cHidden.Bool
	}
	if cCreated.Valid {
		correction.CreatedAt = cCreated.Time
	}
	if cUpdated.Valid {
		correction.UpdatedAt = cUpdated.Time
	}
	target.Correction = correction
}

func (r *Repository) ListDistinctPrimaryTypeDisplayNames() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT trim(primary_type_display_name)
		FROM gemini_direct_spot_google_geocodes
		WHERE primary_type_display_name IS NOT NULL AND trim(primary_type_display_name) <> ''
		ORDER BY trim(primary_type_display_name) ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying distinct primary type display names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning primary type display name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating primary type display names: %w", err)
	}
	return names, nil
}

func (r *Repository) ReplaceSpotCategoryMappings(mappings []SpotCategoryMapping) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin category mapping replacement: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM spot_category_mappings`); err != nil {
		return fmt.Errorf("clearing spot_category_mappings: %w", err)
	}
	stmt, err := tx.Prepare(`
		INSERT INTO spot_category_mappings (primary_type_display_name, category_name, confidence, reason, model, mapped_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return fmt.Errorf("preparing spot_category_mappings insert: %w", err)
	}
	defer stmt.Close()
	for _, mapping := range mappings {
		primary := strings.TrimSpace(mapping.PrimaryTypeDisplayName)
		category := strings.TrimSpace(mapping.CategoryName)
		if primary == "" {
			return fmt.Errorf("spot category mapping primary type display name is required")
		}
		if category == "" {
			return fmt.Errorf("spot category mapping category name is required for %q", primary)
		}
		if _, err := stmt.Exec(primary, category, mapping.Confidence, mapping.Reason, strings.TrimSpace(mapping.Model)); err != nil {
			return fmt.Errorf("inserting spot category mapping %q: %w", primary, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit category mapping replacement: %w", err)
	}
	return nil
}

func (r *Repository) ExportDataForSource(source models.SpotSource) (*models.ExportData, error) {
	normalized, err := models.NormalizeSpotSource(string(source))
	if err != nil {
		return nil, err
	}
	_ = normalized
	return r.exportGeminiDirectData()
}

func (r *Repository) exportGeminiDirectData() (*models.ExportData, error) {
	rows, err := r.db.Query(`
		SELECT
			CAST(gasp.article_source_id AS TEXT) || ':' || CAST(gsm.gemini_direct_spot_mention_id AS TEXT),
			COALESCE(gsg.google_place_id, ''),
			COALESCE(gsm.place, ''),
			NULLIF(p.presenter_name, ''),
			gsg.latitude,
			gsg.longitude,
			COALESCE(scm.category_name, 'Overig'),
			COALESCE(gsm.youtube_url, ''),
			COALESCE(s.url, ''),
			s.published_at,
			gsm.youtube_timestamp_seconds,
			gc.spot_name,
			gc.place_id,
			gc.latitude,
			gc.longitude,
			gc.youtube_timestamp_seconds
		FROM gemini_direct_article_spots gasp
		JOIN gemini_direct_spot_google_geocodes gsg ON gsg.gemini_direct_spot_google_geocode_id = gasp.gemini_direct_spot_google_geocode_id
		JOIN gemini_direct_spot_mentions gsm ON gsm.gemini_direct_spot_mention_id = gsg.gemini_direct_spot_mention_id
		JOIN article_sources s ON s.article_source_id = gasp.article_source_id
		LEFT JOIN (
			SELECT article_source_id, MAX(presenter_id) AS presenter_id
			FROM article_presenters
			GROUP BY article_source_id
		) ap ON ap.article_source_id = gasp.article_source_id
		LEFT JOIN presenters p ON p.presenter_id = ap.presenter_id
		LEFT JOIN gemini_direct_spot_corrections gc ON gc.spot_id = CAST(gasp.article_source_id AS TEXT) || ':' || CAST(gsm.gemini_direct_spot_mention_id AS TEXT)
		LEFT JOIN spot_category_mappings scm ON scm.primary_type_display_name = trim(gsg.primary_type_display_name)
		WHERE COALESCE(gc.hidden, 0) = 0
	`)
	if err != nil {
		return nil, fmt.Errorf("querying Gemini-direct export data: %w", err)
	}
	defer rows.Close()

	data := &models.ExportData{Spots: []models.ExportSpot{}}
	spotPublishedAt := make(map[string]time.Time)
	presenterLatestPublishedAt := make(map[string]time.Time)
	categoryCounts := make(map[string]int)

	for rows.Next() {
		var (
			spot             models.ExportSpot
			rawYouTubeLink   string
			categoryName     string
			publishedAtValue sql.NullString
			sourceTS         sql.NullFloat64
			cSpotName        sql.NullString
			cPlaceID         sql.NullString
			cLat             sql.NullFloat64
			cLng             sql.NullFloat64
			cTimestamp       sql.NullInt64
		)
		var presenterName sql.NullString
		if err := rows.Scan(&spot.SpotID, &spot.PlaceID, &spot.SpotName, &presenterName, &spot.Latitude, &spot.Longitude, &categoryName, &rawYouTubeLink, &spot.ArticleURL, &publishedAtValue, &sourceTS, &cSpotName, &cPlaceID, &cLat, &cLng, &cTimestamp); err != nil {
			return nil, fmt.Errorf("scanning Gemini-direct export row: %w", err)
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
		spot.CategoryName = strings.TrimSpace(categoryName)
		if spot.CategoryName == "" {
			spot.CategoryName = "Overig"
		}
		if cSpotName.Valid {
			spot.SpotName = cSpotName.String
		}
		if cPlaceID.Valid && strings.TrimSpace(cPlaceID.String) != "" {
			if !cLat.Valid || !cLng.Valid {
				return nil, fmt.Errorf("spotId %s has corrected placeId %q without resolved coordinates", spot.SpotID, cPlaceID.String)
			}
			spot.PlaceID = cPlaceID.String
			spot.Latitude = cLat.Float64
			spot.Longitude = cLng.Float64
		}
		if cTimestamp.Valid {
			if cTimestamp.Int64 < 0 {
				return nil, fmt.Errorf("spotId %s has invalid corrected YouTube timestamp %d", spot.SpotID, cTimestamp.Int64)
			}
			spot.YouTubeLink = withYouTubeTimestampSeconds(rawYouTubeLink, cTimestamp.Int64)
		} else {
			spot.YouTubeLink = withYouTubeTimestamp(rawYouTubeLink, sourceTS, sql.NullFloat64{}, sql.NullFloat64{})
		}
		data.Spots = append(data.Spots, spot)
		spotPublishedAt[spot.SpotID] = publishedAt
		if presenterName.Valid {
			name := strings.TrimSpace(presenterName.String)
			if name != "" {
				current, ok := presenterLatestPublishedAt[name]
				if !ok || publishedAt.After(current) {
					presenterLatestPublishedAt[name] = publishedAt
				}
			}
		}
		categoryCounts[spot.CategoryName]++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating Gemini-direct export rows: %w", err)
	}
	slices.SortFunc(data.Spots, func(a, b models.ExportSpot) int {
		at, bt := spotPublishedAt[a.SpotID], spotPublishedAt[b.SpotID]
		if at.Before(bt) {
			return -1
		}
		if bt.Before(at) {
			return 1
		}
		if a.SpotID < b.SpotID {
			return -1
		}
		if a.SpotID > b.SpotID {
			return 1
		}
		return 0
	})

	presenterNames := make([]string, 0, len(presenterLatestPublishedAt))
	for name := range presenterLatestPublishedAt {
		presenterNames = append(presenterNames, name)
	}
	slices.SortFunc(presenterNames, func(a, b string) int {
		at, bt := presenterLatestPublishedAt[a], presenterLatestPublishedAt[b]
		if at.After(bt) {
			return -1
		}
		if bt.After(at) {
			return 1
		}
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})
	for _, name := range presenterNames {
		data.Presenters = append(data.Presenters, models.ExportPresenter{PresenterName: name})
	}

	categoryNames := make([]string, 0, len(categoryCounts))
	for name := range categoryCounts {
		categoryNames = append(categoryNames, name)
	}
	slices.SortFunc(categoryNames, func(a, b string) int {
		if a == "Overig" && b != "Overig" {
			return 1
		}
		if b == "Overig" && a != "Overig" {
			return -1
		}
		if categoryCounts[a] > categoryCounts[b] {
			return -1
		}
		if categoryCounts[a] < categoryCounts[b] {
			return 1
		}
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})
	for _, name := range categoryNames {
		data.Categories = append(data.Categories, models.ExportCategory{CategoryName: name})
	}
	return data, nil
}

func (r *Repository) ResetGeminiDerivedData() error {
	_, err := r.db.Exec(`
		DELETE FROM gemini_direct_spot_corrections;
		DELETE FROM gemini_direct_article_spots;
		DELETE FROM gemini_direct_spot_google_geocodes;
		DELETE FROM gemini_direct_spot_mentions;
	`)
	if err != nil {
		return fmt.Errorf("resetting Gemini-derived data: %w", err)
	}
	return nil
}

func withYouTubeTimestamp(raw string, refinedTS, originalTS, sentenceStartTS sql.NullFloat64) string {
	ts := pickTimestamp(refinedTS, originalTS, sentenceStartTS)
	if ts == nil || *ts < 0 {
		return raw
	}
	return withYouTubeTimestampSeconds(raw, int64(*ts))
}

func withYouTubeTimestampSeconds(raw string, seconds int64) string {
	if seconds < 0 {
		return raw
	}
	videoID := extractYouTubeVideoID(raw)
	if videoID == "" {
		return raw
	}
	return fmt.Sprintf("https://youtu.be/%s?t=%d", videoID, seconds)
}

func pickedTimestampSeconds(refinedTS, originalTS, sentenceStartTS sql.NullFloat64) *int64 {
	ts := pickTimestamp(refinedTS, originalTS, sentenceStartTS)
	if ts == nil || *ts < 0 {
		return nil
	}
	seconds := int64(*ts)
	return &seconds
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
