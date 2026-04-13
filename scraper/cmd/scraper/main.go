package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/scraper"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	dbPath := flag.String("db", filepath.Clean("../data/spots.db"), "path to the SQLite database")
	flag.Parse()

	// Init DB
	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer repo.Close()

	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}
	log.Printf("Database initialized at %s", *dbPath)

	// Phase 1: Crawl category pages for article URLs
	log.Println("Crawling category pages...")
	urls, err := scraper.CrawlArticleURLs()
	if err != nil {
		log.Fatalf("Failed to crawl article URLs: %v", err)
	}
	log.Printf("Found %d article URLs", len(urls))

	// Phase 2: Fetch and store raw HTML
	log.Println("Fetching articles...")
	if err := scraper.FetchAndStoreArticles(urls, repo); err != nil {
		log.Fatalf("Failed to fetch articles: %v", err)
	}
	log.Println("All articles fetched and stored")

	// Phase 3: report pending article count for the upcoming video-based extraction pipeline.
	pending, err := repo.GetPendingArticles()
	if err != nil {
		log.Fatalf("Failed to get pending articles: %v", err)
	}

	log.Printf("Fetch complete. Pending articles in database: %d", len(pending))
	log.Println("Article text LLM extraction has been removed. Next step is to implement video/transcript-based spot extraction from the embedded videos.")
}
