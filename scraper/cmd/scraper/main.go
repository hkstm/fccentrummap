package main

import (
	"context"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hkstm/fccentrummap/internal/repository"
	"github.com/hkstm/fccentrummap/internal/scraper"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	dbPath := flag.String("db", filepath.Clean("../data/spots.db"), "path to the SQLite database")
	articleURL := flag.String("article-url", "", "optional single article URL for focused end-to-end run (must be http(s) on fccentrum.nl)")
	flag.Parse()

	if *articleURL != "" {
		u, err := url.Parse(*articleURL)
		if err != nil {
			log.Fatalf("Invalid -article-url %q: parse error: %v", *articleURL, err)
		}
		host := strings.ToLower(u.Hostname())
		if (u.Scheme != "http" && u.Scheme != "https") || host == "" || (host != "fccentrum.nl" && !strings.HasSuffix(host, ".fccentrum.nl")) {
			log.Fatalf("Invalid -article-url %q: must be an http(s) URL on fccentrum.nl", *articleURL)
		}
	}

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

	// Phase 1: Determine article URLs
	var urls []string
	if *articleURL != "" {
		urls = []string{*articleURL}
		log.Printf("Using single article URL: %s", *articleURL)
	} else {
		log.Println("Crawling category pages...")
		urls, err = scraper.CrawlArticleURLs()
		if err != nil {
			log.Fatalf("Failed to crawl article URLs: %v", err)
		}
		log.Printf("Found %d article URLs", len(urls))
	}

	// Phase 2: Fetch and store raw HTML
	log.Println("Fetching articles...")
	if err := scraper.FetchAndStoreArticles(urls, repo); err != nil {
		log.Fatalf("Failed to fetch articles: %v", err)
	}
	log.Println("All articles fetched and stored")

	// Phase 3: acquire and persist article audio blobs for detected embedded videos.
	log.Println("Acquiring YouTube audio blobs for detected videos...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := scraper.AcquireAndStoreAudio(ctx, repo, nil); err != nil {
		log.Fatalf("Audio acquisition finished with errors: %v", err)
	}
	log.Println("Audio acquisition complete")
}
