package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/pipeline/acquireaudio"
	"github.com/hkstm/fccentrummap/internal/pipeline/collectarticleurls"
	"github.com/hkstm/fccentrummap/internal/pipeline/exportdata"
	"github.com/hkstm/fccentrummap/internal/pipeline/extractspots"
	"github.com/hkstm/fccentrummap/internal/pipeline/fetcharticles"
	"github.com/hkstm/fccentrummap/internal/pipeline/geocodespots"
	"github.com/hkstm/fccentrummap/internal/pipeline/transcribeaudio"
	"github.com/hkstm/fccentrummap/internal/repository"
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
			&cli.StringFlag{Name: "in", Usage: "required for --io file unless --article-url is provided"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("collect-article-urls", mode); err != nil {
				return err
			}
			svc := collectarticleurls.NewService(collectarticleurls.NewSQLiteAdapter(), collectarticleurls.NewFileAdapter())
			res, err := svc.Run(ctx, mode, collectarticleurls.Request{DBPath: cmd.String("db-path"), ArticleURL: cmd.String("article-url"), InputPath: cmd.String("in")})
			if err != nil {
				return err
			}
			if mode == ioSQLite {
				fmt.Printf("seeded %d article urls\n", len(res.URLs))
			} else {
				fmt.Printf("artifact=%s\n", res.OutputPath)
			}
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
			mode := cmd.String("io")
			if err := validateStageMode("fetch-articles", mode); err != nil {
				return err
			}
			svc := fetcharticles.NewService(fetcharticles.NewSQLiteAdapter(), fetcharticles.NewFileAdapter())
			res, err := svc.Run(ctx, mode, fetcharticles.Request{DBPath: cmd.String("db-path"), InputPath: cmd.String("in")})
			if err != nil {
				return err
			}
			if mode == ioSQLite {
				fmt.Printf("fetched %d articles\n", res.FetchedCount)
			} else {
				fmt.Printf("artifact=%s\n", res.OutputPath)
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
			mode := cmd.String("io")
			if err := validateStageMode("acquire-audio", mode); err != nil {
				return err
			}
			svc := acquireaudio.NewService(acquireaudio.NewSQLiteAdapter(), acquireaudio.NewFileAdapter())
			res, err := svc.Run(ctx, mode, acquireaudio.Request{DBPath: cmd.String("db-path"), InputPath: cmd.String("in")})
			if err != nil {
				return err
			}
			if mode == ioSQLite {
				fmt.Println("acquired audio")
			} else {
				fmt.Printf("artifact=%s\n", res.OutputPath)
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
			mode := cmd.String("io")
			if err := validateStageMode("transcribe-audio", mode); err != nil {
				return err
			}
			svc := transcribeaudio.NewService(transcribeaudio.NewSQLiteAdapter(), transcribeaudio.NewFileAdapter())
			res, err := svc.Run(ctx, mode, transcribeaudio.Request{DBPath: cmd.String("db-path"), InputPath: cmd.String("in"), Language: cmd.String("language")})
			if err != nil {
				return err
			}
			if mode == ioSQLite {
				fmt.Printf("transcription_id=%d\n", res.TranscriptionID)
			} else {
				fmt.Printf("artifact=%s\n", res.OutputPath)
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
			mode := cmd.String("io")
			if err := validateStageMode("extract-spots", mode); err != nil {
				return err
			}
			svc := extractspots.NewService(extractspots.NewSQLiteAdapter(), extractspots.NewFileAdapter())
			res, err := svc.Run(ctx, mode, extractspots.Request{DBPath: cmd.String("db-path"), InputPath: cmd.String("in"), OutDir: cmd.String("out-dir"), GemmaModel: cmd.String("gemma-model"), APIKey: defaultGeminiAPIKey(), Endpoint: strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")), UseLatest: true, PersistData: true})
			if err != nil {
				return err
			}
			if mode == ioSQLite {
				fmt.Printf("spot_extraction_id=%d\n", res.SpotExtractionID)
			} else {
				fmt.Printf("artifact=%s\n", res.OutputPath)
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
			mode := cmd.String("io")
			if err := validateStageMode("geocode-spots", mode); err != nil {
				return err
			}
			svc := geocodespots.NewService(geocodespots.NewSQLiteAdapter(), geocodespots.NewFileAdapter())
			res, err := svc.Run(ctx, mode, geocodespots.Request{InputPath: cmd.String("in")})
			if err != nil {
				return err
			}
			fmt.Printf("artifact=%s\n", res.OutputPath)
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
			mode := cmd.String("io")
			if err := validateStageMode("export-data", mode); err != nil {
				return err
			}
			svc := exportdata.NewService(exportdata.NewSQLiteAdapter(), exportdata.NewFileAdapter())
			res, err := svc.Run(ctx, mode, exportdata.Request{DBPath: cmd.String("db-path"), OutputPath: cmd.String("out"), InputPath: cmd.String("in")})
			if err != nil {
				return err
			}
			if mode == ioSQLite {
				fmt.Printf("exported=%s\n", res.OutputPath)
			} else {
				fmt.Printf("artifact=%s\n", res.OutputPath)
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

func fatalf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
	os.Exit(1)
}
