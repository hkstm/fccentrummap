package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/genai"
	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/pipeline/collectarticleurls"
	"github.com/hkstm/fccentrummap/internal/pipeline/exportdata"
	"github.com/hkstm/fccentrummap/internal/pipeline/extractspotsgeminidirect"
	"github.com/hkstm/fccentrummap/internal/pipeline/fetcharticles"
	"github.com/hkstm/fccentrummap/internal/pipeline/geocodespots"
	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/spotcategories"
	"github.com/hkstm/fccentrummap/internal/spotcorrections"
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
			extractSpotsGeminiDirectCommand(),
			geocodeSpotsCommand(),
			mapSpotCategoriesCommand(),
			correctSpotCommand(),
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
			&cli.BoolFlag{Name: "reset-gemini-derived", Value: false, Usage: "delete Gemini-derived mentions/geocodes/links/corrections while preserving articles/fetches"},
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite|file"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = ctx
			ioMode := cmd.String("io")
			if err := validateStageMode("init", ioMode); err != nil {
				return err
			}
			dbPath := strings.TrimSpace(cmd.String("db-path"))
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
			if cmd.Bool("reset-gemini-derived") {
				if err := repo.ResetGeminiDerivedData(); err != nil {
					return err
				}
				fmt.Printf("reset_gemini_derived db=%s\n", dbPath)
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
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite (file not supported yet)"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "article-url", Usage: "optional single article URL"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("collect-article-urls", mode); err != nil {
				return err
			}
			req, err := normalizeCollectArticleURLsRequest(cmd.String("db-path"), cmd.String("article-url"))
			if err != nil {
				return err
			}
			svc := collectarticleurls.NewService(collectarticleurls.NewSQLiteAdapter(), collectarticleurls.NewFileAdapter())
			res, err := svc.Run(ctx, mode, req)
			if err != nil {
				return err
			}
			fmt.Printf("seeded %d article urls\n", len(res.URLs))
			return nil
		},
	}
}

func fetchArticlesCommand() *cli.Command {
	return &cli.Command{
		Name:  "fetch-articles",
		Usage: "Fetch article content for pending article URLs",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite (file not supported yet)"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("fetch-articles", mode); err != nil {
				return err
			}
			req, err := normalizeFetchArticlesRequest(cmd.String("db-path"))
			if err != nil {
				return err
			}
			svc := fetcharticles.NewService(fetcharticles.NewSQLiteAdapter(), fetcharticles.NewFileAdapter())
			res, err := svc.Run(ctx, mode, req)
			if err != nil {
				return err
			}
			fmt.Printf("fetched %d articles\n", res.FetchedCount)
			return nil
		},
	}
}

func extractSpotsGeminiDirectCommand() *cli.Command {
	return &cli.Command{
		Name:  "extract-spots-gemini-direct",
		Usage: "Experimentally extract spot candidates directly from article and YouTube URLs with Gemini",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite (file not supported)"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "out-dir", Value: filepath.Join(cliutil.DefaultDataDir(), "gemini-direct"), Usage: "directory for Gemini-direct artifacts"},
			&cli.StringFlag{Name: "model", Value: defaultGenAIModel(), Usage: "Gemini model identifier"},
			&cli.IntFlag{Name: "limit", Value: 0, Usage: "maximum number of new Gemini requests to make (0 means no limit)"},
			&cli.BoolFlag{Name: "force", Value: false, Usage: "regenerate artifacts and call Gemini even when parsed artifacts already exist"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("extract-spots-gemini-direct", mode); err != nil {
				return err
			}
			req, err := normalizeExtractSpotsGeminiDirectRequest(cmd.String("db-path"), cmd.String("out-dir"), cmd.String("model"), cmd.Bool("force"), cmd.Int("limit"))
			if err != nil {
				return err
			}
			svc := extractspotsgeminidirect.NewService(extractspotsgeminidirect.NewSQLiteAdapter(), extractspotsgeminidirect.NewFileAdapter())
			res, err := svc.Run(ctx, mode, req)
			if err != nil {
				return err
			}
			fmt.Printf("processed=%d skipped=%d cached=%d out_dir=%s\n", res.ProcessedCount, res.SkippedCount, res.CachedCount, res.OutputDir)
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
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database (sqlite mode)"},
			&cli.StringFlag{Name: "in", Usage: "required for --io file"},
			&cli.StringFlag{Name: "spot-source", Value: string(models.SpotSourceGeminiDirect), Usage: "spot source for SQLite mode: gemini-direct"},
			&cli.BoolFlag{Name: "export-json", Value: false, Usage: "also export scraping data JSON after geocoding (sqlite mode only)"},
			&cli.StringFlag{Name: "export-out", Value: filepath.Clean("../viz/public/data/spots.json"), Usage: "JSON export output path"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("geocode-spots", mode); err != nil {
				return err
			}
			req, err := normalizeGeocodeSpotsRequest(mode, cmd.String("db-path"), cmd.String("in"), cmd.String("spot-source"))
			if err != nil {
				return err
			}
			svc := geocodespots.NewService(geocodespots.NewSQLiteAdapter(), geocodespots.NewFileAdapter())
			res, err := svc.Run(ctx, mode, req)
			if err != nil {
				return err
			}
			fmt.Printf("artifact=%s spot_source=%s\n", res.OutputPath, req.SpotSource)

			if cmd.Bool("export-json") {
				if mode != ioSQLite {
					return fmt.Errorf("--export-json requires --io sqlite")
				}
				exportReq, err := normalizeExportDataRequest(cmd.String("db-path"), cmd.String("export-out"), string(req.SpotSource))
				if err != nil {
					return err
				}
				exportSvc := exportdata.NewService(exportdata.NewSQLiteAdapter(), exportdata.NewFileAdapter())
				exportRes, err := exportSvc.Run(ctx, mode, exportReq)
				if err != nil {
					return err
				}
				fmt.Printf("exported=%s\n", exportRes.OutputPath)
			}
			return nil
		},
	}
}

func normalizeCorrectionSpotIDArgument(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("correct-spot requires a stable spotId or map URL containing ?spot=")
	}

	if parsed, err := url.Parse(value); err == nil {
		if spotID := strings.TrimSpace(parsed.Query().Get("spot")); spotID != "" {
			return spotID, nil
		}
	}

	if before, after, ok := strings.Cut(value, "spot="); ok {
		if strings.Contains(before, "://") || strings.Contains(before, "?") || strings.Contains(before, "&") || strings.Contains(before, "#") || before == "" {
			spotID := after
			for _, sep := range []string{"&", "#"} {
				if head, _, found := strings.Cut(spotID, sep); found {
					spotID = head
				}
			}
			decoded, err := url.QueryUnescape(spotID)
			if err != nil {
				return "", fmt.Errorf("decoding spot query parameter: %w", err)
			}
			if decoded = strings.TrimSpace(decoded); decoded != "" {
				return decoded, nil
			}
			return "", fmt.Errorf("correct-spot URL is missing a non-empty spot query parameter")
		}
	}

	if strings.Contains(value, "://") || strings.Contains(value, "?") {
		return "", fmt.Errorf("correct-spot URL is missing a non-empty spot query parameter")
	}

	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return "", fmt.Errorf("decoding spot id: %w", err)
	}
	if decoded = strings.TrimSpace(decoded); decoded != "" {
		return decoded, nil
	}
	return "", fmt.Errorf("correct-spot requires a stable spotId or map URL containing ?spot=")
}

func mapSpotCategoriesCommand() *cli.Command {
	return &cli.Command{
		Name:  "map-spot-categories",
		Usage: "Map observed Google Places primary type display names to frontend categories with Gemini",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite (file not supported)"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "model", Value: defaultGenAIModel(), Usage: "Gemini model identifier"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("map-spot-categories", mode); err != nil {
				return err
			}
			dbPath := strings.TrimSpace(cmd.String("db-path"))
			model := strings.TrimSpace(cmd.String("model"))
			if model == "" {
				model = defaultGenAIModel()
			}
			repo, err := repository.New(dbPath)
			if err != nil {
				return err
			}
			defer repo.Close()
			if err := repo.InitSchema(); err != nil {
				return err
			}
			client := genai.NewClientWithEndpoint(defaultGeminiAPIKey(), model, strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")))
			mapped, err := spotcategories.NewService(repo, client, model).Run(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("mapped=%d model=%s\n", mapped, model)
			return nil
		},
	}
}

func correctSpotCommand() *cli.Command {
	return &cli.Command{
		Name:      "correct-spot",
		Usage:     "Interactively create or update a spot correction by stable spotId or map URL",
		ArgsUsage: "<spot-id-or-url>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "spot-source", Value: string(models.SpotSourceGeminiDirect), Usage: "spot source: gemini-direct"},
			&cli.BoolFlag{Name: "hide", Usage: "hide the spot from exported visualization data"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			spotID, err := normalizeCorrectionSpotIDArgument(cmd.Args().First())
			if err != nil {
				return err
			}
			repo, err := repository.New(strings.TrimSpace(cmd.String("db-path")))
			if err != nil {
				return err
			}
			defer repo.Close()
			if err := repo.InitSchema(); err != nil {
				return err
			}
			source, err := models.NormalizeSpotSource(cmd.String("spot-source"))
			if err != nil {
				return err
			}
			if cmd.Bool("hide") {
				updated, err := spotcorrections.NewService(repo, &lazyGooglePlaceLookup{}).SaveCorrection(ctx, spotcorrections.CorrectionInput{SpotID: spotID, SpotSource: source, Hide: true})
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Hidden spotId %s\n", updated.SpotID)
				return nil
			}
			return runInteractiveSpotCorrection(ctx, repo, &lazyGooglePlaceLookup{}, os.Stdin, os.Stdout, spotID, source)
		},
	}
}

type lazyGooglePlaceLookup struct {
	client *geocoder.Geocoder
}

func (l *lazyGooglePlaceLookup) clientOrNew() (*geocoder.Geocoder, error) {
	if l.client == nil {
		client, err := geocoder.New()
		if err != nil {
			return nil, err
		}
		l.client = client
	}
	return l.client, nil
}

func (l *lazyGooglePlaceLookup) LookupPlaceIDCoordinates(ctx context.Context, placeID string) (*geocoder.Coordinates, error) {
	client, err := l.clientOrNew()
	if err != nil {
		return nil, err
	}
	return client.LookupPlaceIDCoordinates(ctx, placeID)
}

func (l *lazyGooglePlaceLookup) ResolvePlaceInput(ctx context.Context, input string) (*geocoder.Coordinates, error) {
	client, err := l.clientOrNew()
	if err != nil {
		return nil, err
	}
	return client.ResolvePlaceInput(ctx, input)
}

func runInteractiveSpotCorrection(ctx context.Context, repo spotcorrections.Repository, lookup spotcorrections.PlaceIDLookup, in io.Reader, out io.Writer, spotID string, source models.SpotSource) error {
	target, err := repo.GetSpotCorrectionTargetForSource(source, spotID)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("spotId %s was not found among exportable source spots", spotID)
	}

	fmt.Fprintf(out, "Correcting spotId: %s\n", target.SpotID)
	fmt.Fprintf(out, "Current name: %s\n", target.EffectiveSpotName)
	fmt.Fprintf(out, "Current placeId: %s\n", target.EffectivePlaceID)
	if target.EffectiveYouTubeTimestampSecs != nil {
		fmt.Fprintf(out, "Current YouTube timestamp: %ds\n", *target.EffectiveYouTubeTimestampSecs)
	} else {
		fmt.Fprintln(out, "Current YouTube timestamp: none")
	}
	fmt.Fprintln(out, "Leave a prompt blank to preserve the current effective value.")

	scanner := bufio.NewScanner(in)
	readPrompt := func(label string) (string, error) {
		fmt.Fprint(out, label)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		return strings.TrimSpace(scanner.Text()), nil
	}

	name, err := readPrompt("Updated spot name: ")
	if err != nil {
		return err
	}
	placeID, err := readPrompt("Updated Google placeId, Maps URL, or address: ")
	if err != nil {
		return err
	}
	timestampURL, err := readPrompt("Timestamped YouTube URL: ")
	if err != nil {
		return err
	}

	input := spotcorrections.CorrectionInput{SpotID: spotID, SpotSource: source}
	if name != "" {
		input.SpotName = &name
	}
	if placeID != "" {
		input.PlaceID = &placeID
	}
	if timestampURL != "" {
		input.TimestampedYouTubeURL = &timestampURL
	}
	updated, err := spotcorrections.NewService(repo, lookup).SaveCorrection(ctx, input)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Saved correction for spotId %s\n", updated.SpotID)
	return nil
}

func exportDataCommand() *cli.Command {
	return &cli.Command{
		Name:  "export-data",
		Usage: "Export data for visualization",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "io", Value: ioSQLite, Usage: "I/O mode: sqlite (file not supported yet)"},
			&cli.StringFlag{Name: "db-path", Value: cliutil.DefaultDBPath(), Usage: "path to SQLite database"},
			&cli.StringFlag{Name: "out", Value: filepath.Clean("../viz/public/data/spots.json"), Usage: "output path"},
			&cli.StringFlag{Name: "spot-source", Value: string(models.SpotSourceGeminiDirect), Usage: "spot source: gemini-direct"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("io")
			if err := validateStageMode("export-data", mode); err != nil {
				return err
			}
			req, err := normalizeExportDataRequest(cmd.String("db-path"), cmd.String("out"), cmd.String("spot-source"))
			if err != nil {
				return err
			}
			svc := exportdata.NewService(exportdata.NewSQLiteAdapter(), exportdata.NewFileAdapter())
			res, err := svc.Run(ctx, mode, req)
			if err != nil {
				return err
			}
			fmt.Printf("exported=%s spot_source=%s\n", res.OutputPath, req.SpotSource)
			return nil
		},
	}
}

func validateRequiredEnv() error {
	missing := []string{}
	if strings.TrimSpace(os.Getenv("MURMEL_API_KEY")) == "" {
		missing = append(missing, "MURMEL_API_KEY")
	}
	if strings.TrimSpace(os.Getenv("PRODUCTION_GOOGLE_MAPS_API_KEY")) == "" {
		missing = append(missing, "PRODUCTION_GOOGLE_MAPS_API_KEY")
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

func normalizeCollectArticleURLsRequest(dbPath, articleURL string) (collectarticleurls.Request, error) {
	req := collectarticleurls.Request{
		DBPath:     strings.TrimSpace(dbPath),
		ArticleURL: strings.TrimSpace(articleURL),
	}
	return req, nil
}

func normalizeFetchArticlesRequest(dbPath string) (fetcharticles.Request, error) {
	req := fetcharticles.Request{DBPath: strings.TrimSpace(dbPath)}
	return req, nil
}

func normalizeExtractSpotsGeminiDirectRequest(dbPath, outDir, model string, force bool, limit int) (extractspotsgeminidirect.Request, error) {
	req := extractspotsgeminidirect.Request{
		DBPath:   strings.TrimSpace(dbPath),
		OutDir:   strings.TrimSpace(outDir),
		Model:    strings.TrimSpace(model),
		APIKey:   defaultGeminiAPIKey(),
		Endpoint: strings.TrimSpace(os.Getenv("GOOGLE_GENERATIVE_LANGUAGE_ENDPOINT")),
		Force:    force,
		Limit:    limit,
	}
	if req.Limit < 0 {
		return extractspotsgeminidirect.Request{}, fmt.Errorf("--limit must be >= 0")
	}
	if req.OutDir == "" {
		req.OutDir = filepath.Join(cliutil.DefaultDataDir(), "gemini-direct")
	}
	if req.Model == "" {
		req.Model = defaultGenAIModel()
	}
	return req, nil
}

func normalizeGeocodeSpotsRequest(mode, dbPath, inputPath, spotSource string) (geocodespots.Request, error) {
	source, err := models.NormalizeSpotSource(spotSource)
	if err != nil {
		return geocodespots.Request{}, err
	}
	req := geocodespots.Request{DBPath: strings.TrimSpace(dbPath), InputPath: strings.TrimSpace(inputPath), SpotSource: source}
	if mode == ioFile && req.InputPath == "" {
		return geocodespots.Request{}, fmt.Errorf("geocodespots file input requires inputPath")
	}
	return req, nil
}

func normalizeExportDataRequest(dbPath, outputPath, spotSource string) (exportdata.Request, error) {
	source, err := models.NormalizeSpotSource(spotSource)
	if err != nil {
		return exportdata.Request{}, err
	}
	req := exportdata.Request{DBPath: strings.TrimSpace(dbPath), OutputPath: strings.TrimSpace(outputPath), SpotSource: source}
	if req.OutputPath == "" {
		req.OutputPath = filepath.Clean("../viz/public/data/spots.json")
	}
	return req, nil
}

func validateStageMode(stage, mode string) error {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode != ioSQLite && mode != ioFile {
		return fmt.Errorf("invalid --io value %q (expected sqlite|file)", mode)
	}
	supported := map[string]map[string]bool{
		"init":                        {ioSQLite: true},
		"collect-article-urls":        {ioSQLite: true},
		"fetch-articles":              {ioSQLite: true},
		"extract-spots-gemini-direct": {ioSQLite: true},
		"geocode-spots":               {ioSQLite: true, ioFile: true},
		"map-spot-categories":         {ioSQLite: true},
		"export-data":                 {ioSQLite: true},
	}
	if !supported[stage][mode] {
		return fmt.Errorf("stage %s does not support --io %s", stage, mode)
	}
	return nil
}

func defaultGenAIModel() string {
	if model := strings.TrimSpace(os.Getenv("MODEL")); model != "" {
		return model
	}
	return "gemini-3.1-pro-preview"
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
