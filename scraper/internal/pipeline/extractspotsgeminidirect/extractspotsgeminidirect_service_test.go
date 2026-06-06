package extractspotsgeminidirect

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/geminidirect"
	genaiclient "github.com/hkstm/fccentrummap/internal/genai"
	"github.com/hkstm/fccentrummap/internal/repository"
	gogenai "google.golang.org/genai"
	_ "modernc.org/sqlite"
)

type fakeInputLoader struct{ rows []ArticleInput }

func (f fakeInputLoader) ListGeminiDirectInputs() ([]ArticleInput, error) { return f.rows, nil }

type fakeGeminiClient struct {
	calls     int
	lastParts []*gogenai.Part
	body      []byte
}

func (f *fakeGeminiClient) GenerateContentWithParts(ctx context.Context, parts []*gogenai.Part, config *gogenai.GenerateContentConfig) (*genaiclient.GenerateContentResult, error) {
	_ = ctx
	_ = config
	f.calls++
	f.lastParts = parts
	body := f.body
	if len(body) == 0 {
		body = []byte(`{"candidates":[{"content":{"parts":[{"text":"{\"presenter_name\":\"Sam\",\"spots\":[{\"place\":\"Cafe\",\"youtubeTimestampSeconds\":12,\"evidence\":\"article/video\",\"confidence\":0.8,\"notes\":\"\"}]}"}]}}]}`)
	}
	return &genaiclient.GenerateContentResult{StatusCode: 200, Body: body}, nil
}

type fakeAcceptedPersister struct {
	articles []AcceptedArticle
}

func (f *fakeAcceptedPersister) PersistAcceptedGeminiDirectArticle(article AcceptedArticle) error {
	f.articles = append(f.articles, article)
	return nil
}

func TestRunnerSkipsMissingURLAndWritesArtifacts(t *testing.T) {
	outDir := t.TempDir()
	client := &fakeGeminiClient{}
	runner := Runner{
		Inputs: fakeInputLoader{rows: []ArticleInput{
			{ArticleSourceID: 1, ArticleURL: "https://example.test/missing-video"},
			{ArticleSourceID: 2, ArticleURL: "https://example.test/article", YouTubeURL: "https://youtube.com/watch?v=abc"},
		}},
		Client: client,
		Writer: geminidirect.ArtifactWriter{OutDir: outDir},
		Model:  "gemini-test",
	}
	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.SkippedCount != 1 || res.ProcessedCount != 1 || client.calls != 1 {
		t.Fatalf("unexpected counts: skipped=%d processed=%d calls=%d", res.SkippedCount, res.ProcessedCount, client.calls)
	}
	if !strings.Contains(res.Articles[0].Diagnostics[0], "YouTube URL") {
		t.Fatalf("missing explicit diagnostic: %#v", res.Articles[0].Diagnostics)
	}
	for _, suffix := range []string{"prompt.txt", "raw_response.json", "parsed.json"} {
		path := filepath.Join(outDir, "article_2_"+suffix)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
	if len(client.lastParts) != 2 {
		t.Fatalf("expected video and text parts, got %d", len(client.lastParts))
	}
	if client.lastParts[0].FileData == nil || client.lastParts[0].FileData.FileURI != "https://youtube.com/watch?v=abc" || client.lastParts[0].FileData.MIMEType != "video/*" {
		t.Fatalf("expected YouTube video file part, got %#v", client.lastParts[0])
	}
}

func TestRunnerPersistsAcceptedGeminiDirectResponse(t *testing.T) {
	outDir := t.TempDir()
	client := &fakeGeminiClient{}
	persister := &fakeAcceptedPersister{}
	runner := Runner{
		Inputs:    fakeInputLoader{rows: []ArticleInput{{ArticleSourceID: 7, ArticleURL: "https://example.test/article", YouTubeURL: "https://youtube.com/watch?v=abc"}}},
		Client:    client,
		Writer:    geminidirect.ArtifactWriter{OutDir: outDir},
		Persister: persister,
		Model:     "gemini-test",
	}
	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ProcessedCount != 1 || len(persister.articles) != 1 {
		t.Fatalf("expected one persisted article, res=%+v persisted=%+v", res, persister.articles)
	}
	article := persister.articles[0]
	if article.ArticleSourceID != 7 || article.PresenterName == nil || *article.PresenterName != "Sam" || len(article.Spots) != 1 {
		t.Fatalf("unexpected persisted article: %+v", article)
	}
	if article.Spots[0].Place != "Cafe" || article.Spots[0].Confidence == nil || *article.Spots[0].Confidence != 0.8 {
		t.Fatalf("unexpected persisted spot: %+v", article.Spots[0])
	}
}

func TestRunnerRejectsInvalidGeminiContextBeforePersistence(t *testing.T) {
	client := &fakeGeminiClient{body: []byte(`{"article_url":"https://wrong.test/article","youtube_url":"https://youtube.com/watch?v=abc","presenter_name":"Sam","spots":[{"place":"Cafe","youtubeTimestampSeconds":12,"evidence":"article/video","confidence":0.8}]}`)}
	persister := &fakeAcceptedPersister{}
	runner := Runner{
		Inputs:    fakeInputLoader{rows: []ArticleInput{{ArticleSourceID: 7, ArticleURL: "https://example.test/article", YouTubeURL: "https://youtube.com/watch?v=abc"}}},
		Client:    client,
		Writer:    geminidirect.ArtifactWriter{OutDir: t.TempDir()},
		Persister: persister,
		Model:     "gemini-test",
	}
	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run should preserve diagnostics without hard failure: %v", err)
	}
	if len(persister.articles) != 0 {
		t.Fatalf("invalid context was persisted: %+v", persister.articles)
	}
	if len(res.Articles) != 1 || len(res.Articles[0].Diagnostics) == 0 || !strings.Contains(res.Articles[0].Diagnostics[0], "article_url") {
		t.Fatalf("expected context diagnostic, got %+v", res.Articles)
	}
}

func TestRunnerLimitCapsNewGeminiRequests(t *testing.T) {
	outDir := t.TempDir()
	client := &fakeGeminiClient{}
	runner := Runner{
		Inputs: fakeInputLoader{rows: []ArticleInput{
			{ArticleSourceID: 1, ArticleURL: "https://example.test/one", YouTubeURL: "https://youtube.com/watch?v=one"},
			{ArticleSourceID: 2, ArticleURL: "https://example.test/two", YouTubeURL: "https://youtube.com/watch?v=two"},
		}},
		Client: client,
		Writer: geminidirect.ArtifactWriter{OutDir: outDir},
		Model:  "gemini-test",
		Limit:  1,
	}
	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if client.calls != 1 || res.ProcessedCount != 1 || len(res.Articles) != 1 {
		t.Fatalf("expected exactly one attempted article, calls=%d res=%+v", client.calls, res)
	}
	if _, err := os.Stat(filepath.Join(outDir, "article_1_parsed.json")); err != nil {
		t.Fatalf("expected first parsed artifact: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "article_2_parsed.json")); !os.IsNotExist(err) {
		t.Fatalf("second parsed artifact should not exist, stat err=%v", err)
	}
}

func TestRunnerSkipsExistingParsedArtifactUnlessForced(t *testing.T) {
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "article_1_parsed.json"), []byte(`{"articleUrl":"https://example.test/article","youtubeUrl":"https://youtube.com/watch?v=abc","model":"cached","spots":[]}`), 0o644); err != nil {
		t.Fatalf("seed parsed artifact: %v", err)
	}
	client := &fakeGeminiClient{}
	runner := Runner{
		Inputs: fakeInputLoader{rows: []ArticleInput{{ArticleSourceID: 1, ArticleURL: "https://example.test/article", YouTubeURL: "https://youtube.com/watch?v=abc"}}},
		Client: client,
		Writer: geminidirect.ArtifactWriter{OutDir: outDir},
		Model:  "gemini-test",
	}
	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run cached: %v", err)
	}
	if client.calls != 0 || res.CachedCount != 1 || res.SkippedCount != 1 || !res.Articles[0].Cached {
		t.Fatalf("expected cached skip without Gemini call, calls=%d res=%+v", client.calls, res)
	}

	runner.Force = true
	res, err = runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run forced: %v", err)
	}
	if client.calls != 1 || res.CachedCount != 0 || res.ProcessedCount != 1 {
		t.Fatalf("expected forced regeneration, calls=%d res=%+v", client.calls, res)
	}
}

func TestSQLiteInputDoesNotRequireLegacyTablesAndRunnerDoesNotMutateCorrections(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("repository.New: %v", err)
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	articleID, err := repo.UpsertArticleSource("https://example.test/article")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if _, err := repo.UpsertArticleFetch(articleID, `<html><meta property="article:published_time" content="2024-01-01T00:00:00Z"><iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ"></iframe></html>`); err != nil {
		t.Fatalf("UpsertArticleFetch: %v", err)
	}

	before := canonicalCounts(t, dbPath)
	client := &fakeGeminiClient{}
	runner := Runner{Inputs: repositoryInputLoader{repo: repo}, Client: client, Writer: geminidirect.ArtifactWriter{OutDir: t.TempDir()}, Model: "gemini-test"}
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	after := canonicalCounts(t, dbPath)
	if before != after {
		t.Fatalf("canonical table counts mutated: before=%v after=%v", before, after)
	}
	if client.calls != 1 {
		t.Fatalf("expected one Gemini call from URL-only input, got %d", client.calls)
	}
}

type canonicalTableCounts struct{ corrections int }

func canonicalCounts(t *testing.T, dbPath string) canonicalTableCounts {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	return canonicalTableCounts{
		corrections: countTable(t, db, "gemini_direct_spot_corrections"),
	}
}

func countTable(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}
