package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/models"
	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/scraper"
)

func main() {
	log.SetFlags(0)

	dbPath := flag.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database (SPOTS_DB_PATH overrides default)")
	limit := flag.Int("limit", 5, "number of most recent articles to process")
	dryRun := flag.Bool("dry-run", false, "compute extraction outcomes without writing to database")
	printContent := flag.Bool("print-content", false, "print extracted text segments for visual verification")
	outFile := flag.String("out-file", "", "optional path to write extracted text report")
	flag.Parse()

	if *limit <= 0 {
		log.Fatalf("invalid --limit %d: must be > 0", *limit)
	}

	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	if err := repo.InitSchema(); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	articles, err := repo.ListRecentArticles(*limit)
	if err != nil {
		log.Fatalf("failed to load recent articles: %v", err)
	}
	if len(articles) == 0 {
		fmt.Println("No articles found.")
		return
	}

	matched := 0
	noMatch := 0
	errored := 0
	reportLines := make([]string, 0)

	for _, article := range articles {
		result := scraper.ExtractArticleTextContent(article.HTML)
		result.ArticleRawID = article.ArticleRawID

		if !*dryRun {
			if err := repo.ReplaceArticleTextExtraction(result); err != nil {
				log.Fatalf("failed to persist extraction article_raw_id=%d: %v", article.ArticleRawID, err)
			}
		}

		switch result.Status {
		case models.ArticleTextExtractionStatusMatched:
			matched++
		case models.ArticleTextExtractionStatusNoMatch:
			noMatch++
		default:
			errored++
		}

		errText := ""
		if result.ErrorMessage != nil {
			errText = fmt.Sprintf(" error=%q", *result.ErrorMessage)
		}

		action := "persisted"
		if *dryRun {
			action = "dry-run"
		}
		fmt.Printf("[%s] article_raw_id=%d status=%s mode=%s matched_count=%d url=%s%s\n",
			action,
			article.ArticleRawID,
			result.Status,
			result.ExtractionMode,
			result.MatchedCount,
			article.URL,
			errText,
		)
		if *printContent {
			for i, content := range result.Contents {
				fmt.Printf("  - [%d] (%s) %s\n", i+1, content.SourceType, content.Content)
			}
		}

		if *outFile != "" {
			reportLines = append(reportLines,
				fmt.Sprintf("# article_raw_id=%d", article.ArticleRawID),
				fmt.Sprintf("url: %s", article.URL),
				fmt.Sprintf("status: %s", result.Status),
				fmt.Sprintf("mode: %s", result.ExtractionMode),
				fmt.Sprintf("matched_count: %d", result.MatchedCount),
			)
			if result.ErrorMessage != nil {
				reportLines = append(reportLines, fmt.Sprintf("error: %s", *result.ErrorMessage))
			}
			reportLines = append(reportLines, "content:")
			if len(result.Contents) == 0 {
				reportLines = append(reportLines, "  (none)")
			} else {
				for i, content := range result.Contents {
					reportLines = append(reportLines, fmt.Sprintf("  [%d] (%s) %s", i+1, content.SourceType, content.Content))
				}
			}
			reportLines = append(reportLines, "")
		}
	}

	if *outFile != "" {
		reportPath := filepath.Clean(*outFile)
		if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
			log.Fatalf("failed to create output directory for %s: %v", reportPath, err)
		}
		if err := os.WriteFile(reportPath, []byte(strings.Join(reportLines, "\n")), 0o644); err != nil {
			log.Fatalf("failed to write output report %s: %v", reportPath, err)
		}
		fmt.Printf("report_file=%s\n", reportPath)
	}

	fmt.Println()
	if *dryRun {
		fmt.Printf("Summary (dry-run): processed=%d matched=%d no_match=%d error=%d\n", len(articles), matched, noMatch, errored)
	} else {
		fmt.Printf("Summary: processed=%d matched=%d no_match=%d error=%d\n", len(articles), matched, noMatch, errored)
	}
}
