package extraction

import "testing"

func TestTwoPassBatchedExtractionFlowMultiPlace(t *testing.T) {
	pass1Raw := []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_spots", "args": {
		"presenter_name": "Niels Oosthoek",
		"spots": [
		  {"place": "Stopera", "sentenceStartTimestamp": 15.0},
		  {"place": "Oosterpark", "sentenceStartTimestamp": 44.0}
		]
	  }}}]}}]
	}`)
	pass2Raw := []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_refined_spots", "args": {
		"spots": [
		  {"place": "Stopera", "refinedSentenceStartTimestamp": 12.0},
		  {"place": "Oosterpark", "refinedSentenceStartTimestamp": 39.5}
		]
	  }}}]}}]
	}`)

	pass1, err := ParseAndValidateResponse(pass1Raw)
	if err != nil {
		t.Fatalf("parse pass1: %v", err)
	}
	pass2, err := ParseAndValidateRefinementResponse(pass2Raw, pass1.Spots)
	if err != nil {
		t.Fatalf("parse pass2: %v", err)
	}
	final := ApplyRefinements(pass1, pass2)
	if len(final.Spots) != 2 {
		t.Fatalf("expected 2 spots, got %d", len(final.Spots))
	}
	if *final.Spots[0].OriginalSentenceStartTimestamp != 15.0 || *final.Spots[0].RefinedSentenceStartTimestamp != 12.0 {
		t.Fatalf("unexpected first spot refinement: %+v", final.Spots[0])
	}
	if *final.Spots[1].OriginalSentenceStartTimestamp != 44.0 || *final.Spots[1].RefinedSentenceStartTimestamp != 39.5 {
		t.Fatalf("unexpected second spot refinement: %+v", final.Spots[1])
	}
	if *final.Spots[0].SentenceStartTimestamp != 12.0 || *final.Spots[1].SentenceStartTimestamp != 39.5 {
		t.Fatalf("expected refined timestamp to be primary downstream timestamp: %+v", final.Spots)
	}
}
