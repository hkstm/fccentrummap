package extraction

import "testing"

func TestParseAndValidateResponseFromFunctionCallArgs(t *testing.T) {
	raw := []byte(`{
	  "candidates": [
	    {
	      "content": {
	        "parts": [
	          {
	            "functionCall": {
	              "name": "submit_spots",
	              "args": {
	                "presenter_name": "Niels Oosthoek",
	                "spots": [
	                  {"place": "Stopera", "sentenceStartTimestamp": 15.13}
	                ]
	              }
	            }
	          }
	        ]
	      }
	    }
	  ]
	}`)

	parsed, err := ParseAndValidateResponse(raw)
	if err != nil {
		t.Fatalf("ParseAndValidateResponse() error = %v", err)
	}
	if parsed.PresenterName == nil || *parsed.PresenterName != "Niels Oosthoek" {
		t.Fatalf("unexpected presenter_name: %+v", parsed.PresenterName)
	}
	if len(parsed.Spots) != 1 {
		t.Fatalf("expected 1 spot, got %d", len(parsed.Spots))
	}
	if parsed.Spots[0].Place != "Stopera" {
		t.Fatalf("unexpected place: %+v", parsed.Spots[0])
	}
	if parsed.Spots[0].OriginalSentenceStartTimestamp == nil || *parsed.Spots[0].OriginalSentenceStartTimestamp != 15.13 {
		t.Fatalf("unexpected originalSentenceStartTimestamp: %+v", parsed.Spots[0].OriginalSentenceStartTimestamp)
	}
	if parsed.Spots[0].RefinedSentenceStartTimestamp == nil || *parsed.Spots[0].RefinedSentenceStartTimestamp != 15.13 {
		t.Fatalf("unexpected refinedSentenceStartTimestamp: %+v", parsed.Spots[0].RefinedSentenceStartTimestamp)
	}
}

func TestParseAndValidateResponseRejectsMissingPresenterName(t *testing.T) {
	raw := []byte(`{
	  "candidates": [
	    {
	      "content": {
	        "parts": [
	          {
	            "functionCall": {
	              "name": "submit_spots",
	              "args": {
	                "spots": [
	                  {"place": "Stopera", "sentenceStartTimestamp": 15.13}
	                ]
	              }
	            }
	          }
	        ]
	      }
	    }
	  ]
	}`)

	_, err := ParseAndValidateResponse(raw)
	if err == nil {
		t.Fatal("expected error for missing presenter_name")
	}
}

func TestParseAndValidateResponseRejectsMissingTimestamp(t *testing.T) {
	raw := []byte(`{
	  "candidates": [
	    {
	      "content": {
	        "parts": [
	          {
	            "functionCall": {
	              "name": "submit_spots",
	              "args": {
	                "spots": [
	                  {"place": "Stopera"}
	                ]
	              }
	            }
	          }
	        ]
	      }
	    }
	  ]
	}`)

	_, err := ParseAndValidateResponse(raw)
	if err == nil {
		t.Fatal("expected error for missing sentenceStartTimestamp")
	}
}

func TestParseAndValidateResponseRejectsDuplicatePlaces(t *testing.T) {
	raw := []byte(`{
	  "candidates": [
	    {
	      "content": {
	        "parts": [
	          {
	            "functionCall": {
	              "name": "submit_spots",
	              "args": {
	                "spots": [
	                  {"place": " Stopera ", "sentenceStartTimestamp": 15.13},
	                  {"place": "Stopera", "sentenceStartTimestamp": 18.0}
	                ]
	              }
	            }
	          }
	        ]
	      }
	    }
	  ]
	}`)

	_, err := ParseAndValidateResponse(raw)
	if err == nil {
		t.Fatal("expected duplicate place validation error")
	}
}

func TestParseAndValidateRefinementResponseRejectsUnknownPlace(t *testing.T) {
	orig := 15.13
	pass1 := []Candidate{{Place: "Stopera", OriginalSentenceStartTimestamp: &orig}}
	raw := []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_refined_spots", "args": {
		"spots": [{"place":"Onbekend", "refinedSentenceStartTimestamp": 10.0}]
	  }}}]}}]
	}`)

	_, err := ParseAndValidateRefinementResponse(raw, pass1)
	if err == nil {
		t.Fatal("expected unknown place validation error")
	}
}

func TestParseAndValidateRefinementResponseRejectsLaterTimestamp(t *testing.T) {
	p1 := 15.0
	pass1 := []Candidate{{Place: "Stopera", OriginalSentenceStartTimestamp: &p1, SentenceStartTimestamp: &p1}}
	raw := []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_refined_spots", "args": {
		"spots": [{"place":"Stopera", "refinedSentenceStartTimestamp": 16.0}]
	  }}}]}}]
	}`)

	_, err := ParseAndValidateRefinementResponse(raw, pass1)
	if err == nil {
		t.Fatal("expected later-than-pass1 refinement validation error")
	}
}

func TestParseAndValidateRefinementResponseRejectsPreviousPass1BoundViolation(t *testing.T) {
	p1, p2 := 15.0, 40.0
	pass1 := []Candidate{
		{Place: "Stopera", OriginalSentenceStartTimestamp: &p1, SentenceStartTimestamp: &p1},
		{Place: "Oosterpark", OriginalSentenceStartTimestamp: &p2, SentenceStartTimestamp: &p2},
	}
	raw := []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_refined_spots", "args": {
		"spots": [
		  {"place":"Stopera", "refinedSentenceStartTimestamp": 14.0},
		  {"place":"Oosterpark", "refinedSentenceStartTimestamp": 14.5}
		]
	  }}}]}}]
	}`)

	_, err := ParseAndValidateRefinementResponse(raw, pass1)
	if err == nil {
		t.Fatal("expected previous pass-1 ordering validation error")
	}
}

func TestApplyRefinementsValidationRules(t *testing.T) {
	p1 := 15.0
	p2 := 40.0
	pass1 := &ParsedResponse{Spots: []Candidate{
		{Place: "Stopera", SentenceStartTimestamp: &p1, OriginalSentenceStartTimestamp: &p1, RefinedSentenceStartTimestamp: &p1},
		{Place: "Oosterpark", SentenceStartTimestamp: &p2, OriginalSentenceStartTimestamp: &p2, RefinedSentenceStartTimestamp: &p2},
	}}

	t.Run("earlier acceptance", func(t *testing.T) {
		r1, r2 := 12.0, 39.0
		out := ApplyRefinements(pass1, &ParsedRefinementResponse{Spots: []RefinementCandidate{
			{Place: "Stopera", RefinedSentenceStartTimestamp: &r1},
			{Place: "Oosterpark", RefinedSentenceStartTimestamp: &r2},
		}})
		if *out.Spots[0].RefinedSentenceStartTimestamp != 12.0 || *out.Spots[1].RefinedSentenceStartTimestamp != 39.0 {
			t.Fatalf("unexpected refined timestamps: %+v", out.Spots)
		}
	})

	t.Run("equal no-op acceptance", func(t *testing.T) {
		r1, r2 := 15.0, 40.0
		out := ApplyRefinements(pass1, &ParsedRefinementResponse{Spots: []RefinementCandidate{
			{Place: "Stopera", RefinedSentenceStartTimestamp: &r1},
			{Place: "Oosterpark", RefinedSentenceStartTimestamp: &r2},
		}})
		if *out.Spots[0].RefinedSentenceStartTimestamp != 15.0 || *out.Spots[1].RefinedSentenceStartTimestamp != 40.0 {
			t.Fatalf("expected no-op refinement acceptance, got %+v", out.Spots)
		}
	})

	t.Run("later rejection", func(t *testing.T) {
		r1 := 16.0
		out := ApplyRefinements(pass1, &ParsedRefinementResponse{Spots: []RefinementCandidate{{Place: "Stopera", RefinedSentenceStartTimestamp: &r1}}})
		if *out.Spots[0].RefinedSentenceStartTimestamp != 15.0 {
			t.Fatalf("expected original timestamp due to later rejection, got %v", *out.Spots[0].RefinedSentenceStartTimestamp)
		}
	})

	t.Run("previous-place-bound rejection", func(t *testing.T) {
		r1, r2 := 14.0, 13.0
		out := ApplyRefinements(pass1, &ParsedRefinementResponse{Spots: []RefinementCandidate{
			{Place: "Stopera", RefinedSentenceStartTimestamp: &r1},
			{Place: "Oosterpark", RefinedSentenceStartTimestamp: &r2},
		}})
		if *out.Spots[1].RefinedSentenceStartTimestamp != 40.0 {
			t.Fatalf("expected original second timestamp due to previous-place bound rejection, got %v", *out.Spots[1].RefinedSentenceStartTimestamp)
		}
	})
}
