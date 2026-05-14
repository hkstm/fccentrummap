package contracttests

import "testing"

func TestWriterOwnershipMappingIsOneWriterPerTable(t *testing.T) {
	ownership := map[string][]string{
		"article_sources":       {"collect-article-urls"},
		"article_fetches":       {"fetch-articles"},
		"article_texts":         {"extract-article-text"},
		"audio_sources":         {"acquire-audio"},
		"audio_transcriptions":  {"transcribe-audio"},
		"spot_mentions":         {"extract-spots"},
		"presenters":            {"extract-spots"},
		"article_presenters":    {"extract-spots"},
		"spot_google_geocodes":  {"geocode-spots"},
		"article_spots":         {"geocode-spots"},
	}

	for table, writers := range ownership {
		if len(writers) != 1 {
			t.Fatalf("table %s must have exactly one writer stage, got %v", table, writers)
		}
		if writers[0] == "" {
			t.Fatalf("table %s has empty writer stage", table)
		}
	}
}
