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
	if parsed.Spots[0].SentenceStartTimestamp == nil || *parsed.Spots[0].SentenceStartTimestamp != 15.13 {
		t.Fatalf("unexpected sentenceStartTimestamp: %+v", parsed.Spots[0].SentenceStartTimestamp)
	}
}

func TestParseAndValidateResponseAllowsMissingPresenterName(t *testing.T) {
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
	if parsed.PresenterName != nil {
		t.Fatalf("expected presenter_name=nil, got %+v", parsed.PresenterName)
	}
}

func TestParseAndValidateResponsePresenterNameVariants(t *testing.T) {
	cases := []struct {
		name      string
		raw       []byte
		expectNil bool
	}{
		{
			name: "null",
			raw: []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_spots", "args": {"presenter_name": null, "spots": [{"place": "Stopera", "sentenceStartTimestamp": 15.13}]}}}]}}]
	}`),
			expectNil: true,
		},
		{
			name: "whitespace",
			raw: []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_spots", "args": {"presenter_name": "   ", "spots": [{"place": "Stopera", "sentenceStartTimestamp": 15.13}]}}}]}}]
	}`),
			expectNil: true,
		},
		{
			name: "present",
			raw: []byte(`{
	  "candidates": [{"content": {"parts": [{"functionCall": {"name": "submit_spots", "args": {"presenter_name": "Niels", "spots": [{"place": "Stopera", "sentenceStartTimestamp": 15.13}]}}}]}}]
	}`),
			expectNil: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseAndValidateResponse(tc.raw)
			if err != nil {
				t.Fatalf("ParseAndValidateResponse() error = %v", err)
			}
			if tc.expectNil && parsed.PresenterName != nil {
				t.Fatalf("expected presenter_name=nil, got %+v", parsed.PresenterName)
			}
			if !tc.expectNil && (parsed.PresenterName == nil || *parsed.PresenterName != "Niels") {
				t.Fatalf("expected presenter_name=Niels, got %+v", parsed.PresenterName)
			}
		})
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
