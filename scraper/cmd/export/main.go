package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/hkstm/fccentrummap/internal/repository"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	dbPath := flag.String("db", filepath.Clean("../data/spots.db"), "path to the SQLite database")
	outPath := flag.String("out", filepath.Clean("../viz/public/data/spots.json"), "path to the output JSON file")
	flag.Parse()

	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer repo.Close()

	data, err := repo.ExportData()
	if err != nil {
		log.Fatalf("Failed to export data: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	file, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to write JSON: %v", err)
	}

	log.Printf("Exported %d spots for %d authors to %s", len(data.Spots), len(data.Authors), *outPath)
}
