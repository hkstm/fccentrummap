package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/extractspots"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/scraper"
	"github.com/hkstm/fccentrummap/internal/transcription"
)

const (
	ioSQLite = "sqlite"
	ioFile   = "file"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		fatalf("usage: scrape <init|collect-article-urls|fetch-articles|acquire-audio|transcribe-audio|extract-spots|geocode-spots|export-data> [flags]")
	}

	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "collect-article-urls":
		runCollectArticleURLs(os.Args[2:])
	case "fetch-articles":
		runFetchArticles(os.Args[2:])
	case "acquire-audio":
		runAcquireAudio(os.Args[2:])
	case "transcribe-audio":
		runTranscribeAudio(os.Args[2:])
	case "extract-spots":
		runExtractSpots(os.Args[2:])
	case "geocode-spots":
		runGeocodeSpots(os.Args[2:])
	case "export-data":
		runExportData(os.Args[2:])
	default:
		fatalf("unknown subcommand: %s", os.Args[1])
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	reset := fs.Bool("reset", false, "remove database file before schema init")
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	fs.Parse(args)
	validateStageModeOrDie("init", *ioMode)
	if err := validateRequiredEnv(); err != nil {
		fatalf("init preflight failed: %v", err)
	}
	if *reset {
		if err := os.Remove(*dbPath); err != nil && !os.IsNotExist(err) {
			fatalf("failed to reset database %s: %v", *dbPath, err)
		}
	}
	repo := mustRepo(*dbPath)
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		fatalf("failed to initialize schema: %v", err)
	}
	fmt.Printf("initialized db=%s\n", *dbPath)
}

func runCollectArticleURLs(args []string) {
	fs := flag.NewFlagSet("collect-article-urls", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	articleURL := fs.String("article-url", "", "optional single article URL")
	inPath := fs.String("in", "", "required for --io file")
	fs.Parse(args)
	validateStageModeOrDie("collect-article-urls", *ioMode)

	if *ioMode == ioSQLite {
		repo := mustRepo(*dbPath)
		defer repo.Close()
		must(repo.InitSchema())
		if strings.TrimSpace(*articleURL) != "" {
			must(repo.InsertArticleRaw(strings.TrimSpace(*articleURL), "", nil))
			fmt.Printf("seeded article url=%s\n", strings.TrimSpace(*articleURL))
			return
		}
		urls, err := scraper.CrawlArticleURLs()
		must(err)
		for _, u := range urls {
			must(repo.InsertArticleRaw(u, "", nil))
		}
		fmt.Printf("seeded %d article urls\n", len(urls))
		return
	}

	if strings.TrimSpace(*inPath) == "" && strings.TrimSpace(*articleURL) == "" {
		fatalf("collect-article-urls --io file requires --in or --article-url")
	}
	identity := strings.TrimSpace(*articleURL)
	if identity == "" {
		identity = filepath.Base(strings.TrimSpace(*inPath))
	}
	out := cliutil.StageArtifactPath(cliutil.DefaultDataDir(), "collect-article-urls", identity, "articles", "json")
	must(os.MkdirAll(filepath.Dir(out), 0o755))
	payload := map[string]any{"articleUrl": strings.TrimSpace(*articleURL), "in": strings.TrimSpace(*inPath)}
	b, _ := json.MarshalIndent(payload, "", "  ")
	must(os.WriteFile(out, b, 0o644))
	fmt.Printf("artifact=%s\n", out)
}

func runFetchArticles(args []string) {
	fs := flag.NewFlagSet("fetch-articles", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	inPath := fs.String("in", "", "required for --io file")
	fs.Parse(args)
	validateStageModeOrDie("fetch-articles", *ioMode)

	if *ioMode == ioSQLite {
		repo := mustRepo(*dbPath)
		defer repo.Close()
		must(repo.InitSchema())
		articles, err := repo.GetPendingArticles()
		must(err)
		urls := make([]string, 0, len(articles))
		for _, a := range articles {
			urls = append(urls, a.URL)
		}
		must(scraper.FetchAndStoreArticles(urls, repo))
		fmt.Printf("fetched %d articles\n", len(urls))
		return
	}
	if strings.TrimSpace(*inPath) == "" {
		fatalf("fetch-articles --io file requires --in")
	}
	writePassthroughArtifact("fetch-articles", *inPath, "fetched")
}

func runAcquireAudio(args []string) {
	fs := flag.NewFlagSet("acquire-audio", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	inPath := fs.String("in", "", "required for --io file")
	fs.Parse(args)
	validateStageModeOrDie("acquire-audio", *ioMode)
	if *ioMode == ioSQLite {
		repo := mustRepo(*dbPath)
		defer repo.Close()
		must(repo.InitSchema())
		must(scraper.AcquireAndStoreAudio(context.Background(), repo, nil))
		fmt.Println("acquired audio")
		return
	}
	if strings.TrimSpace(*inPath) == "" {
		fatalf("acquire-audio --io file requires --in")
	}
	writePassthroughArtifact("acquire-audio", *inPath, "audio")
}

func runTranscribeAudio(args []string) {
	fs := flag.NewFlagSet("transcribe-audio", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	inPath := fs.String("in", "", "required for --io file")
	language := fs.String("language", "nl", "language code sent to Murmel")
	fs.Parse(args)
	validateStageModeOrDie("transcribe-audio", *ioMode)

	if *ioMode == ioSQLite {
		repo := mustRepo(*dbPath)
		defer repo.Close()
		must(repo.InitSchema())
		src, err := repo.GetLatestArticleAudioSource()
		must(err)
		if src == nil {
			fatalf("no audio source rows with non-empty audio_blob found")
		}
		client := transcription.NewMurmelClient(os.Getenv("MURMEL_API_KEY"))
		must(client.Validate())
		filename := fmt.Sprintf("article_audio_source_%d.%s", src.AudioSourceID, cliutil.SafeExt(src.AudioFormat))
		res, err := client.Transcribe(context.Background(), filename, src.AudioBlob, *language)
		must(err)
		msg, jsonErr := canonicalizeJSON(res.Body)
		if jsonErr != nil {
			msg = "{}"
			errMessage := fmt.Sprintf("non-JSON response persisted with fallback payload: %v", jsonErr)
			if res.ErrMessage != nil {
				errMessage = *res.ErrMessage + "; " + errMessage
			}
			res.ErrMessage = &errMessage
		}
		id, err := repo.UpsertArticleAudioTranscription(models.ArticleAudioTranscription{
			AudioSourceID:    src.AudioSourceID,
			Provider:         "murmel",
			Language:         *language,
			HTTPStatus:       res.HTTPStatus,
			ResponseJSON:     msg,
			ResponseByteSize: int64(len(msg)),
			ErrorMessage:     res.ErrMessage,
		})
		must(err)
		fmt.Printf("transcription_id=%d\n", id)
		return
	}
	if strings.TrimSpace(*inPath) == "" {
		fatalf("transcribe-audio --io file requires --in")
	}
	writePassthroughArtifact("transcribe-audio", *inPath, "transcription")
}

func runExtractSpots(args []string) {
	fs := flag.NewFlagSet("extract-spots", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	inPath := fs.String("in", "", "required for --io file")
	outDir := fs.String("out-dir", cliutil.DefaultDataDir(), "directory for extraction artifacts")
	gemmaModel := fs.String("gemma-model", defaultGemmaModel(), "Gemma model identifier")
	apiKey := fs.String("gemini-api-key", defaultGeminiAPIKey(), "Gemini API key")
	endpoint := fs.String("google-endpoint", strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")), "optional endpoint override")
	fs.Parse(args)
	validateStageModeOrDie("extract-spots", *ioMode)
	if *ioMode == ioSQLite {
		repo := mustRepo(*dbPath)
		defer repo.Close()
		must(repo.InitSchema())
		res, err := extractspots.Run(context.Background(), repo, extractspots.Options{
			UseLatest:     true,
			OutDir:        *outDir,
			GemmaModel:    *gemmaModel,
			APIKey:        *apiKey,
			Endpoint:      *endpoint,
			PersistRecord: true,
		})
		must(err)
		fmt.Printf("spot_extraction_id=%d\n", res.SpotExtractionID)
		return
	}
	if strings.TrimSpace(*inPath) == "" {
		fatalf("extract-spots --io file requires --in")
	}
	writePassthroughArtifact("extract-spots", *inPath, "candidates")
}

func runGeocodeSpots(args []string) {
	fs := flag.NewFlagSet("geocode-spots", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	inPath := fs.String("in", "", "required for --io file")
	fs.Parse(args)
	validateStageModeOrDie("geocode-spots", *ioMode)
	if *ioMode == ioSQLite {
		fatalf("geocode-spots does not support --io sqlite yet; sqlite persistence is deferred. Use --io file --in <path>")
	}
	if strings.TrimSpace(*inPath) == "" {
		fatalf("geocode-spots --io file requires --in")
	}

	b, err := os.ReadFile(strings.TrimSpace(*inPath))
	must(err)
	var payload map[string]any
	if err := json.Unmarshal(b, &payload); err != nil {
		fatalf("invalid geocode input JSON: %v", err)
	}
	query, ok := payload["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		fatalf("geocode input missing required field: query")
	}
	query = strings.TrimSpace(query)
	g, err := geocoder.New()
	if err != nil {
		fatalf("failed to initialize geocoder: %v", err)
	}
	coords, err := g.GeocodePlace(context.Background(), query)
	must(err)
	inBase := strings.TrimSuffix(filepath.Base(strings.TrimSpace(*inPath)), filepath.Ext(strings.TrimSpace(*inPath)))
	identity := strings.SplitN(inBase, "__", 2)[0]
	if strings.TrimSpace(identity) == "" {
		fatalf("unable to derive artifact identity from --in path: %s", strings.TrimSpace(*inPath))
	}
	out := cliutil.StageArtifactPath(cliutil.DefaultDataDir(), "geocode-spots", identity, "geocoded", "json")
	must(os.MkdirAll(filepath.Dir(out), 0o755))
	resp := map[string]any{"query": query, "coordinates": coords}
	outBytes, _ := json.MarshalIndent(resp, "", "  ")
	must(os.WriteFile(out, outBytes, 0o644))
	fmt.Printf("artifact=%s\n", out)
}

func runExportData(args []string) {
	fs := flag.NewFlagSet("export-data", flag.ExitOnError)
	ioMode := fs.String("io", ioSQLite, "I/O mode: sqlite|file")
	dbPath := fs.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database")
	outPath := fs.String("out", filepath.Clean("../viz/public/data/spots.json"), "output path")
	inPath := fs.String("in", "", "required for --io file")
	fs.Parse(args)
	validateStageModeOrDie("export-data", *ioMode)
	if *ioMode == ioSQLite {
		repo := mustRepo(*dbPath)
		defer repo.Close()
		data, err := repo.ExportData()
		must(err)
		must(os.MkdirAll(filepath.Dir(*outPath), 0o755))
		bytes, _ := json.MarshalIndent(data, "", "  ")
		must(os.WriteFile(*outPath, bytes, 0o644))
		fmt.Printf("exported=%s\n", *outPath)
		return
	}
	if strings.TrimSpace(*inPath) == "" {
		fatalf("export-data --io file requires --in")
	}
	writePassthroughArtifact("export-data", *inPath, "export")
}

func validateRequiredEnv() error {
	missing := []string{}
	if strings.TrimSpace(os.Getenv("MURMEL_API_KEY")) == "" {
		missing = append(missing, "MURMEL_API_KEY")
	}
	if strings.TrimSpace(os.Getenv("GOOGLE_MAPS_API_KEY")) == "" {
		missing = append(missing, "GOOGLE_MAPS_API_KEY")
	}
	hasGeminiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")) != "" || strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")) != "" || strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_API_KEY")) != ""
	if !hasGeminiKey {
		missing = append(missing, "one of GEMINI_API_KEY | GOOGLE_API_KEY | GOOGLE_GENERATIVE_LANGUAGE_API_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateStageModeOrDie(stage, mode string) {
	if err := validateStageMode(stage, mode); err != nil {
		fatalf("%v", err)
	}
}

func validateStageMode(stage, mode string) error {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode != ioSQLite && mode != ioFile {
		return fmt.Errorf("invalid --io value %q (expected sqlite|file)", mode)
	}
	supported := map[string]map[string]bool{
		"init":                 {ioSQLite: true},
		"collect-article-urls": {ioSQLite: true, ioFile: true},
		"fetch-articles":       {ioSQLite: true, ioFile: true},
		"acquire-audio":        {ioSQLite: true, ioFile: true},
		"transcribe-audio":     {ioSQLite: true, ioFile: true},
		"extract-spots":        {ioSQLite: true, ioFile: true},
		"geocode-spots":        {ioFile: true},
		"export-data":          {ioSQLite: true, ioFile: true},
	}
	if !supported[stage][mode] {
		if stage == "geocode-spots" && mode == ioSQLite {
			return fmt.Errorf("geocode-spots does not support --io sqlite yet; use --io file --in <path>")
		}
		return fmt.Errorf("stage %s does not support --io %s", stage, mode)
	}
	return nil
}

func writePassthroughArtifact(stage, inPath, payloadType string) {
	identity := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
	out := cliutil.StageArtifactPath(cliutil.DefaultDataDir(), stage, identity, payloadType, "json")
	must(os.MkdirAll(filepath.Dir(out), 0o755))
	b, err := os.ReadFile(inPath)
	must(err)
	must(os.WriteFile(out, b, 0o644))
	fmt.Printf("artifact=%s\n", out)
}

func defaultGemmaModel() string {
	if model := strings.TrimSpace(os.Getenv("GEMMA_MODEL")); model != "" {
		return model
	}
	return "gemma-4-31b-it"
}

func defaultGeminiAPIKey() string {
	if v := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_API_KEY")); v != "" {
		return v
	}
	return ""
}

func canonicalizeJSON(raw []byte) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return "", fmt.Errorf("empty response body")
	}
	if !json.Valid(raw) {
		return "", fmt.Errorf("response body is not valid JSON")
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return "", fmt.Errorf("canonicalizing JSON: %w", err)
	}
	return buf.String(), nil
}

func mustRepo(path string) *repository.Repository {
	repo, err := repository.New(path)
	must(err)
	return repo
}

func must(err error) {
	if err != nil {
		fatalf("%v", err)
	}
}

func fatalf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
	os.Exit(1)
}
