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
	if len(parsed.Spots) != 1 {
		t.Fatalf("expected 1 spot, got %d", len(parsed.Spots))
	}
	if parsed.Spots[0].Place != "Stopera" {
		t.Fatalf("unexpected place: %+v", parsed.Spots[0])
	}
	if parsed.Spots[0].SentenceStartTimestamp == nil || *parsed.Spots[0].SentenceStartTimestamp != 15.13 {
		t.Fatalf("unexpected sentenceStartTimestamp: %+v", parsed.Spots[0].SentenceStartTimestamp)
	}
}

func TestParseAndValidateResponseRejectsJSONTextOnlyResponse(t *testing.T) {
	raw := []byte(`{
	  "candidates": [
	    {
	      "content": {
	        "parts": [
	          {"text": "thinking..."},
	          {"text": "{\"spots\":[{\"place\":\"Casa del gusto\",\"sentenceStartTimestamp\":214.72}]}"}
	        ]
	      }
	    }
	  ]
	}`)

	_, err := ParseAndValidateResponse(raw)
	if err == nil {
		t.Fatal("expected error for response without extraction function call output")
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
