DB ?= data/spots.db
EXPORT ?= viz/public/data/spots.json

.PHONY: scrape export build check test viz help

help:
	@echo "Targets: scrape export build test check viz"

scrape:
	cd scraper && go run ./cmd/scraper -db ../$(DB)

export:
	cd scraper && go run ./cmd/export -db ../$(DB) -out ../$(EXPORT)

build:
	cd scraper && go build ./...

test:
	cd scraper && go test ./...

check: test build

viz:
	@echo "Frontend area: viz/"
	@echo "Runtime data contract: $(EXPORT)"
	@echo "Future frontend work should consume generated JSON, not SQLite directly."
