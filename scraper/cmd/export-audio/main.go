package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hkstm/fccentrummap/internal/cliutil"
	"github.com/hkstm/fccentrummap/internal/repository"
)

func main() {
	log.SetFlags(0)

	dbPath := flag.String("db-path", cliutil.DefaultDBPath(), "path to SQLite database (SPOTS_DB_PATH overrides default)")
	audioSourceID := flag.Int64("audio-source-id", 0, "audio_source_id to export")
	outDir := flag.String("out-dir", filepath.Clean("../data"), "directory for exported file")
	flag.Parse()

	if *audioSourceID <= 0 {
		log.Fatalf("--audio-source-id must be greater than 0")
	}

	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	src, err := repo.GetArticleAudioSourceByID(*audioSourceID)
	if err != nil {
		log.Fatalf("failed to load audio source %d: %v", *audioSourceID, err)
	}
	if src == nil {
		log.Fatalf("audio source %d not found", *audioSourceID)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("failed to create output directory %s: %v", *outDir, err)
	}

	fileName := fmt.Sprintf("article_audio_source_%d.%s", src.AudioSourceID, cliutil.SafeExt(src.AudioFormat))
	outPath := filepath.Join(*outDir, fileName)

	if err := os.WriteFile(outPath, src.AudioBlob, 0o600); err != nil {
		log.Fatalf("failed to write output file %s: %v", outPath, err)
	}

	fmt.Printf("exported_audio=%s\n", outPath)
	fmt.Printf("bytes=%d\n", len(src.AudioBlob))
}
