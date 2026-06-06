package contracttests

import "testing"

func TestWriterOwnershipMappingDocumentsSourceSpecificTables(t *testing.T) {
	ownership := map[string][]string{
		"article_sources":                    {"collect-article-urls"},
		"article_fetches":                    {"fetch-articles"},
		"presenters":                         {"extract-spots-gemini-direct"},
		"article_presenters":                 {"extract-spots-gemini-direct"},
		"gemini_direct_spot_mentions":        {"extract-spots-gemini-direct"},
		"gemini_direct_spot_google_geocodes": {"geocode-spots"},
		"gemini_direct_article_spots":        {"geocode-spots"},
		"gemini_direct_spot_corrections":     {"correct-spot"},
	}

	for table, writers := range ownership {
		if len(writers) == 0 {
			t.Fatalf("table %s must document at least one writer stage", table)
		}
		for _, writer := range writers {
			if writer == "" {
				t.Fatalf("table %s has empty writer stage", table)
			}
		}
	}
}
