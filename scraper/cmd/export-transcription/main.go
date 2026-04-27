package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hkstm/fccentrummap/internal/repository"
)

func main() {
	log.SetFlags(0)

	dbPath := flag.String("db-path", filepath.Clean("../data/spots.db"), "path to SQLite database")
	transcriptionID := flag.Int64("transcription-id", 0, "transcription_id to export")
	outDir := flag.String("out-dir", filepath.Clean("../data"), "directory for exported file")
	flag.Parse()

	if *transcriptionID <= 0 {
		log.Fatalf("--transcription-id must be greater than 0")
	}

	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	row, err := repo.GetArticleAudioTranscriptionByID(*transcriptionID)
	if err != nil {
		log.Fatalf("failed to load transcription %d: %v", *transcriptionID, err)
	}
	if row == nil {
		log.Fatalf("transcription %d not found", *transcriptionID)
	}

	if !json.Valid([]byte(row.ResponseJSON)) {
		log.Fatalf("stored transcription %d has invalid JSON payload", *transcriptionID)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("failed to create output directory %s: %v", *outDir, err)
	}

	fileName := fmt.Sprintf("article_audio_transcription_%d.json", row.TranscriptionID)
	outPath := filepath.Join(*outDir, fileName)
	if err := os.WriteFile(outPath, []byte(row.ResponseJSON), 0o644); err != nil {
		log.Fatalf("failed to write output file %s: %v", outPath, err)
	}

	fmt.Printf("exported_transcription=%s\n", outPath)
	fmt.Printf("bytes=%d\n", len(row.ResponseJSON))
}
