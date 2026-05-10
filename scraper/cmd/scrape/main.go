package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/extractspots"
	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/scraper"
	"github.com/hkstm/fccentrummap/internal/transcription"
	"github.com/urfave/cli/v3"
)

const (
	ioSQLite = "sqlite"
	ioFile   = "file"
)

func main() {
	log.SetFlags(0)

	app := &cli.Command{
		Name:  "scrape",
		Usage: "Run FCCentrum scraper pipeline stages",
		Commands: []*cli.Command{
			initCommand(),
			collectArticleURLsCommand(),
			fetchArticlesCommand(),
			acquireAudioCommand(),
			transcribeAudioCommand(),
			extractSpotsCommand(),
			geocodeSpotsCommand(),
			exportDataCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fatalf("%v", err)
	}
}

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize SQLite schema and validate required environment",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.BoolFlag{Name: "reset", Value: false, Usage: "remove database file before schema init"},
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = ctx
			ioMode := cmd.String("io")
			if err := validateStageMode("init", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			if err := validateRequiredEnv(); err != nil {
				return fmt.Errorf("init preflight failed: %w", err)
			}
			if cmd.Bool("reset") {
				if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to reset database %s: %w", dbPath, err)
				}
			}
			repo, err := repository.New(dbPath)
			if err != nil {
				return err
			}
			defer repo.Close()
			if err := repo.InitSchema(); err != nil {
				return fmt.Errorf("failed to initialize schema: %w", err)
			}
			fmt.Printf("initialized db=%s\n", dbPath)
			return nil
		},
	}
}

func collectArticleURLsCommand() *cli.Command {
	return &cli.Command{
		Name:  "collect-article-urls",
		Usage: "Collect and store article URLs",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "article-url", Usage: "optional single article URL"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = ctx
			ioMode := cmd.String("io")
			if err := validateStageMode("collect-article-urls", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			articleURL := cmd.String("article-url")
			inPath := cmd.String("in")

			if ioMode == ioSQLite {
				repo, err := repository.New(dbPath)
				if err != nil {
					return err
				}
				defer repo.Close()
				if err := repo.InitSchema(); err != nil {
					return err
				}
				if strings.TrimSpace(articleURL) != "" {
					if err := repo.InsertArticleRaw(strings.TrimSpace(articleURL), "", nil); err != nil {
						return err
					}
					fmt.Printf("seeded article url=%s\n", strings.TrimSpace(articleURL))
					return nil
				}
				urls, err := scraper.CrawlArticleURLs()
				if err != nil {
					return err
				}
				for _, u := range urls {
					if err := repo.InsertArticleRaw(u, "", nil); err != nil {
						return err
					}
				}
				fmt.Printf("seeded %d article urls\n", len(urls))
				return nil
			}

			if strings.TrimSpace(inPath) == "" && strings.TrimSpace(articleURL) == "" {
				return fmt.Errorf("collect-article-urls --io file requires --in or --article-url")
			}
			identity := strings.TrimSpace(articleURL)
			if identity == "" {
				identity = filepath.Base(strings.TrimSpace(inPath))
			}
			out := cliutil.StageArtifactPath(cliutil.DefaultDataDir(), "collect-article-urls", identity, "articles", "json")
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			payload := map[string]any{"articleUrl": strings.TrimSpace(articleURL), "in": strings.TrimSpace(inPath)}
			b, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(out, b, 0o644); err != nil {
				return err
			}
			fmt.Printf("artifact=%s\n", out)
			return nil
		},
	}
}

func fetchArticlesCommand() *cli.Command {
	return &cli.Command{
		Name:  "fetch-articles",
		Usage: "Fetch article content for pending article URLs",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = ctx
			ioMode := cmd.String("io")
			if err := validateStageMode("fetch-articles", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			inPath := cmd.String("in")

			if ioMode == ioSQLite {
				repo, err := repository.New(dbPath)
				if err != nil {
					return err
				}
				defer repo.Close()
				if err := repo.InitSchema(); err != nil {
					return err
				}
				articles, err := repo.GetPendingArticles()
				if err != nil {
					return err
				}
				urls := make([]string, 0, len(articles))
				for _, a := range articles {
					urls = append(urls, a.URL)
				}
				if err := scraper.FetchAndStoreArticles(urls, repo); err != nil {
					return err
				}
				fmt.Printf("fetched %d articles\n", len(urls))
				return nil
			}
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("fetch-articles --io file requires --in")
			}
			if err := writePassthroughArtifact("fetch-articles", inPath, "fetched"); err != nil {
				return err
			}
			return nil
		},
	}
}

func acquireAudioCommand() *cli.Command {
	return &cli.Command{
		Name:  "acquire-audio",
		Usage: "Acquire and store audio",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			ioMode := cmd.String("io")
			if err := validateStageMode("acquire-audio", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			inPath := cmd.String("in")
			if ioMode == ioSQLite {
				repo, err := repository.New(dbPath)
				if err != nil {
					return err
				}
				defer repo.Close()
				if err := repo.InitSchema(); err != nil {
					return err
				}
				if err := scraper.AcquireAndStoreAudio(ctx, repo, nil); err != nil {
					return err
				}
				fmt.Println("acquired audio")
				return nil
			}
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("acquire-audio --io file requires --in")
			}
			if err := writePassthroughArtifact("acquire-audio", inPath, "audio"); err != nil {
				return err
			}
			return nil
		},
	}
}

func transcribeAudioCommand() *cli.Command {
	return &cli.Command{
		Name:  "transcribe-audio",
		Usage: "Transcribe audio via Murmel",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
			&cli.StringFlag{Name: "language", Value: "nl", Usage: "language code sent to Murmel"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			ioMode := cmd.String("io")
			if err := validateStageMode("transcribe-audio", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			inPath := cmd.String("in")
			language := cmd.String("language")

			if ioMode == ioSQLite {
				repo, err := repository.New(dbPath)
				if err != nil {
					return err
				}
				defer repo.Close()
				if err := repo.InitSchema(); err != nil {
					return err
				}
				src, err := repo.GetLatestArticleAudioSource()
				if err != nil {
					return err
				}
				if src == nil {
					return fmt.Errorf("no audio source rows with non-empty audio_blob found")
				}
				client := transcription.NewMurmelClient(os.Getenv("MURMEL_API_KEY"))
				if err := client.Validate(); err != nil {
					return err
				}
				filename := fmt.Sprintf("article_audio_source_%d.%s", src.AudioSourceID, cliutil.SafeExt(src.AudioFormat))
				res, err := client.Transcribe(ctx, filename, src.AudioBlob, language)
				if err != nil {
					return err
				}
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
					Language:         language,
					HTTPStatus:       res.HTTPStatus,
					ResponseJSON:     msg,
					ResponseByteSize: int64(len(msg)),
					ErrorMessage:     res.ErrMessage,
				})
				if err != nil {
					return err
				}
				fmt.Printf("transcription_id=%d\n", id)
				return nil
			}
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("transcribe-audio --io file requires --in")
			}
			if err := writePassthroughArtifact("transcribe-audio", inPath, "transcription"); err != nil {
				return err
			}
			return nil
		},
	}
}

func extractSpotsCommand() *cli.Command {
	return &cli.Command{
		Name:  "extract-spots",
		Usage: "Extract place candidates from transcription",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
			&cli.StringFlag{Name: "out-dir", Value: cliutil.DefaultDataDir(), Usage: "directory for extraction artifacts"},
			&cli.StringFlag{Name: "gemma-model", Value: defaultGemmaModel(), Usage: "Gemma model identifier"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			ioMode := cmd.String("io")
			if err := validateStageMode("extract-spots", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			inPath := cmd.String("in")
			if ioMode == ioSQLite {
				repo, err := repository.New(dbPath)
				if err != nil {
					return err
				}
				defer repo.Close()
				if err := repo.InitSchema(); err != nil {
					return err
				}
				res, err := extractspots.Run(ctx, repo, extractspots.Options{
					UseLatest:     true,
					OutDir:        cmd.String("out-dir"),
					GemmaModel:    cmd.String("gemma-model"),
					APIKey:        defaultGeminiAPIKey(),
					Endpoint:      strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")),
					PersistRecord: true,
				})
				if err != nil {
					return err
				}
				fmt.Printf("spot_extraction_id=%d\n", res.SpotExtractionID)
				return nil
			}
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("extract-spots --io file requires --in")
			}
			if err := writePassthroughArtifact("extract-spots", inPath, "candidates"); err != nil {
				return err
			}
			return nil
		},
	}
}

func geocodeSpotsCommand() *cli.Command {
	return &cli.Command{
		Name:  "geocode-spots",
		Usage: "Geocode extracted place candidates",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			ioMode := cmd.String("io")
			if err := validateStageMode("geocode-spots", ioMode); err != nil {
				return err
			}
			inPath := cmd.String("in")
			if ioMode == ioSQLite {
				return fmt.Errorf("geocode-spots does not support --io sqlite yet; sqlite persistence is deferred. Use --io file --in <path>")
			}
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("geocode-spots --io file requires --in")
			}

			b, err := os.ReadFile(strings.TrimSpace(inPath))
			if err != nil {
				return err
			}
			var payload map[string]any
			if err := json.Unmarshal(b, &payload); err != nil {
				return fmt.Errorf("invalid geocode input JSON: %w", err)
			}
			query, ok := payload["query"].(string)
			if !ok || strings.TrimSpace(query) == "" {
				return fmt.Errorf("geocode input missing required field: query")
			}
			query = strings.TrimSpace(query)
			g, err := geocoder.New()
			if err != nil {
				return fmt.Errorf("failed to initialize geocoder: %w", err)
			}
			coords, err := g.GeocodePlace(ctx, query)
			if err != nil {
				return err
			}
			inBase := strings.TrimSuffix(filepath.Base(strings.TrimSpace(inPath)), filepath.Ext(strings.TrimSpace(inPath)))
			identity := strings.SplitN(inBase, "__", 2)[0]
			if strings.TrimSpace(identity) == "" {
				return fmt.Errorf("unable to derive artifact identity from --in path: %s", strings.TrimSpace(inPath))
			}
			out := cliutil.StageArtifactPath(cliutil.DefaultDataDir(), "geocode-spots", identity, "geocoded", "json")
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			resp := map[string]any{"query": query, "coordinates": coords}
			outBytes, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(out, outBytes, 0o644); err != nil {
				return err
			}
			fmt.Printf("artifact=%s\n", out)
			return nil
		},
	}
}

func exportDataCommand() *cli.Command {
	return &cli.Command{
		Name:  "export-data",
		Usage: "Export data for visualization",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "out", Value: filepath.Clean("../viz/public/data/spots.json"), Usage: "output path"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = ctx
			ioMode := cmd.String("io")
			if err := validateStageMode("export-data", ioMode); err != nil {
				return err
			}
			dbPath := cmd.String("db-path")
			outPath := cmd.String("out")
			inPath := cmd.String("in")
			if ioMode == ioSQLite {
				repo, err := repository.New(dbPath)
				if err != nil {
					return err
				}
				defer repo.Close()
				data, err := repo.ExportData()
				if err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
					return err
				}
				payload, err := json.MarshalIndent(data, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(outPath, payload, 0o644); err != nil {
					return err
				}
				fmt.Printf("exported=%s\n", outPath)
				return nil
			}
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("export-data --io file requires --in")
			}
			if err := writePassthroughArtifact("export-data", inPath, "export"); err != nil {
				return err
			}
			return nil
		},
	}
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

func writePassthroughArtifact(stage, inPath, payloadType string) error {
	identity := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
	out := cliutil.StageArtifactPath(cliutil.DefaultDataDir(), stage, identity, payloadType, "json")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	b, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, b, 0o644); err != nil {
		return err
	}
	fmt.Printf("artifact=%s\n", out)
	return nil
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

func fatalf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
	os.Exit(1)
}
