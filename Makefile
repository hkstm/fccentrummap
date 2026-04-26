DB ?= data/spots.db
EXPORT ?= viz/public/data/spots.json

.PHONY: scrape export build check test viz setup-hooks help

help:
	@echo "Targets: scrape export build test check viz setup-hooks"

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

setup-hooks:
	@git rev-parse --is-inside-work-tree >/dev/null 2>&1 || { echo "Not inside a git working tree"; exit 1; }
	chmod -R +x .githooks
	git config --local core.hooksPath .githooks
	@echo "Configured git hooks path: .githooks"
