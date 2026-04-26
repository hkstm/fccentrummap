package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	CREATE TABLE IF NOT EXISTS articles_raw (
		article_raw_id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		html TEXT NOT NULL,
		video_id TEXT,
		status TEXT NOT NULL DEFAULT 'PENDING',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS authors (
		author_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS spots (
		spot_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		address TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		UNIQUE(name, address)
	);

	CREATE TABLE IF NOT EXISTS articles (
		article_id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_raw_id INTEGER NOT NULL REFERENCES articles_raw(article_raw_id),
		author_id INTEGER NOT NULL REFERENCES authors(author_id),
		title TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS article_spots (
		article_id INTEGER NOT NULL REFERENCES articles(article_id),
		spot_id INTEGER NOT NULL REFERENCES spots(spot_id),
		PRIMARY KEY (article_id, spot_id)
	);

	CREATE TABLE IF NOT EXISTS article_audio_sources (
		audio_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_raw_id INTEGER NOT NULL UNIQUE REFERENCES articles_raw(article_raw_id),
		video_id TEXT NOT NULL,
		youtube_url TEXT NOT NULL,
		audio_format TEXT NOT NULL,
		mime_type TEXT NOT NULL,
		audio_blob BLOB NOT NULL,
		byte_size INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := r.db.Exec(schema); err != nil {
		return fmt.Errorf("initializing schema: %w", err)
	}

	return nil
}

func (r *Repository) InsertArticleRaw(url, html string, videoID *string) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO articles_raw (url, html, video_id) VALUES (?, ?, ?)`,
		url, html, videoID,
	)
	if err != nil {
		return fmt.Errorf("inserting article_raw url=%s: %w", url, err)
	}
	return nil
}

func (r *Repository) GetPendingArticles() ([]models.ArticleRaw, error) {
	rows, err := r.db.Query(
		`SELECT article_raw_id, url, html, video_id FROM articles_raw WHERE status = 'PENDING'`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending articles: %w", err)
	}
	defer rows.Close()

	var articles []models.ArticleRaw
	for rows.Next() {
		var a models.ArticleRaw
		if err := rows.Scan(&a.ArticleRawID, &a.URL, &a.HTML, &a.VideoID); err != nil {
			return nil, fmt.Errorf("scanning article_raw row: %w", err)
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func (r *Repository) ArticleAudioSourceExists(articleRawID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM article_audio_sources WHERE article_raw_id = ?)`,
		articleRawID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking audio source existence article_raw_id=%d: %w", articleRawID, err)
	}
	return exists, nil
}

func (r *Repository) InsertArticleAudioSource(src models.ArticleAudioSource) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO article_audio_sources
			(article_raw_id, video_id, youtube_url, audio_format, mime_type, audio_blob, byte_size)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		src.ArticleRawID,
		src.VideoID,
		src.YouTubeURL,
		src.AudioFormat,
		src.MIMEType,
		src.AudioBlob,
		src.ByteSize,
	)
	if err != nil {
		return fmt.Errorf("inserting article_audio_source article_raw_id=%d: %w", src.ArticleRawID, err)
	}
	return nil
}

func (r *Repository) GetArticleAudioSource(articleRawID int64) (*models.ArticleAudioSource, error) {
	var src models.ArticleAudioSource
	err := r.db.QueryRow(
		`SELECT audio_source_id, article_raw_id, video_id, youtube_url, audio_format, mime_type, audio_blob, byte_size, created_at
		 FROM article_audio_sources
		 WHERE article_raw_id = ?`,
		articleRawID,
	).Scan(
		&src.AudioSourceID,
		&src.ArticleRawID,
		&src.VideoID,
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
		return nil, fmt.Errorf("querying article_audio_source article_raw_id=%d: %w", articleRawID, err)
	}
	return &src, nil
}

func (r *Repository) InsertAuthor(name string) (int64, error) {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO authors (name) VALUES (?)`,
		name,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting author name=%s: %w", name, err)
	}

	var id int64
	err = r.db.QueryRow(`SELECT author_id FROM authors WHERE name = ?`, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("querying author_id for name=%s: %w", name, err)
	}
	return id, nil
}

func (r *Repository) InsertSpot(name, address string, lat, lng float64) (int64, error) {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO spots (name, address, latitude, longitude) VALUES (?, ?, ?, ?)`,
		name, address, lat, lng,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting spot name=%s address=%s: %w", name, address, err)
	}

	var id int64
	err = r.db.QueryRow(
		`SELECT spot_id FROM spots WHERE name = ? AND address = ?`,
		name, address,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("querying spot_id for name=%s address=%s: %w", name, address, err)
	}
	return id, nil
}

func (r *Repository) InsertArticle(articleRawID, authorID int64, title string) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO articles (article_raw_id, author_id, title) VALUES (?, ?, ?)`,
		articleRawID, authorID, title,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting article raw_id=%d author_id=%d: %w", articleRawID, authorID, err)
	}
	return result.LastInsertId()
}

func (r *Repository) LinkArticleSpots(articleID int64, spotIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("starting article_spots transaction article_id=%d: %w", articleID, err)
	}
	defer tx.Rollback()

	for _, spotID := range spotIDs {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO article_spots (article_id, spot_id) VALUES (?, ?)`,
			articleID, spotID,
		)
		if err != nil {
			return fmt.Errorf("linking article_id=%d spot_id=%d: %w", articleID, spotID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing article_spots transaction article_id=%d: %w", articleID, err)
	}
	return nil
}

func (r *Repository) UpdateArticleRawStatus(articleRawID int64, url, status, reason string) error {
	_, err := r.db.Exec(
		`UPDATE articles_raw SET status = ?, updated_at = ? WHERE article_raw_id = ?`,
		status, time.Now().UTC(), articleRawID,
	)
	if err != nil {
		return fmt.Errorf("updating article_raw status id=%d: %w", articleRawID, err)
	}
	if status == "FAILED" {
		log.Printf("ERROR: article_raw_id=%d url=%s reason=%s", articleRawID, url, reason)
	}
	return nil
}

func (r *Repository) ExportData() (*models.ExportData, error) {
	rows, err := r.db.Query(`
		SELECT s.name, s.address, s.latitude, s.longitude, a.name
		FROM spots s
		JOIN article_spots aps ON aps.spot_id = s.spot_id
		JOIN articles ar ON ar.article_id = aps.article_id
		JOIN authors a ON a.author_id = ar.author_id
		ORDER BY s.name, s.address, a.name
	`)
	if err != nil {
		return nil, fmt.Errorf("querying export data: %w", err)
	}
	defer rows.Close()

	data := &models.ExportData{}
	authorSet := make(map[string]struct{})
	spotIndex := make(map[string]int)

	for rows.Next() {
		var (
			name    string
			address string
			lat     float64
			lng     float64
			author  string
		)
		if err := rows.Scan(&name, &address, &lat, &lng, &author); err != nil {
			return nil, fmt.Errorf("scanning export row: %w", err)
		}

		if _, ok := authorSet[author]; !ok {
			authorSet[author] = struct{}{}
			data.Authors = append(data.Authors, author)
		}

		key := name + "\x00" + address
		idx, ok := spotIndex[key]
		if !ok {
			data.Spots = append(data.Spots, models.ExportSpot{
				Name:    name,
				Address: address,
				Lat:     lat,
				Lng:     lng,
				Authors: []string{},
			})
			idx = len(data.Spots) - 1
			spotIndex[key] = idx
		}
		spot := &data.Spots[idx]
		if len(spot.Authors) == 0 || spot.Authors[len(spot.Authors)-1] != author {
			spot.Authors = append(spot.Authors, author)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating export rows: %w", err)
	}

	return data, nil
}
