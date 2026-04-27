package extraction

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Candidate struct {
	Place                  string   `json:"place"`
	SentenceStartTimestamp *float64 `json:"sentenceStartTimestamp"`
}

type ParsedResponse struct {
	PresenterName *string     `json:"presenter_name"`
	Spots         []Candidate `json:"spots"`
}

type generateContentEnvelope struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text         string `json:"text"`
				Thought      bool   `json:"thought"`
				FunctionCall *struct {
					Name string          `json:"name"`
					Args json.RawMessage `json:"args"`
				} `json:"functionCall"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func ParseAndValidateResponse(rawBody []byte) (*ParsedResponse, error) {
	parsed, err := extractParsedResponse(rawBody)
	if err != nil {
		return nil, err
	}

	for i, item := range parsed.Spots {
		if strings.TrimSpace(item.Place) == "" {
			return nil, fmt.Errorf("model output validation failed: spots[%d].place is required", i)
		}
		if item.SentenceStartTimestamp == nil {
			return nil, fmt.Errorf("model output validation failed: spots[%d].sentenceStartTimestamp is required", i)
		}
	}

	return parsed, nil
}

func extractParsedResponse(rawBody []byte) (*ParsedResponse, error) {
	var envelope generateContentEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return nil, fmt.Errorf("model response is not valid generateContent JSON: %w", err)
	}

	for _, cand := range envelope.Candidates {
		for _, part := range cand.Content.Parts {
			if part.FunctionCall == nil {
				continue
			}
			if part.FunctionCall.Name != SubmitSpotsFunctionName {
				continue
			}
			if len(part.FunctionCall.Args) == 0 {
				return nil, fmt.Errorf("model function call args are empty")
			}
			var parsed ParsedResponse
			if err := json.Unmarshal(part.FunctionCall.Args, &parsed); err != nil {
				return nil, fmt.Errorf("model function call args are not valid extraction JSON: %w", err)
			}
			if parsed.PresenterName != nil {
				trimmed := strings.TrimSpace(*parsed.PresenterName)
				if trimmed == "" {
					parsed.PresenterName = nil
				} else {
					parsed.PresenterName = &trimmed
				}
			}
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("model response does not contain extraction function call output")
}
