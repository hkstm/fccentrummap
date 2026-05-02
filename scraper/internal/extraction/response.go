package extraction

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Candidate struct {
	Place                          string   `json:"place"`
	SentenceStartTimestamp         *float64 `json:"sentenceStartTimestamp,omitempty"`
	OriginalSentenceStartTimestamp *float64 `json:"originalSentenceStartTimestamp,omitempty"`
	RefinedSentenceStartTimestamp  *float64 `json:"refinedSentenceStartTimestamp,omitempty"`
}

type ParsedResponse struct {
	PresenterName *string     `json:"presenter_name,omitempty"`
	Spots         []Candidate `json:"spots"`
}

type RefinementCandidate struct {
	Place                         string   `json:"place"`
	RefinedSentenceStartTimestamp *float64 `json:"refinedSentenceStartTimestamp"`
}

type ParsedRefinementResponse struct {
	Spots []RefinementCandidate `json:"spots"`
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
	var parsed ParsedResponse
	if err := extractFunctionArgs(rawBody, SubmitSpotsFunctionName, &parsed); err != nil {
		return nil, err
	}

	if parsed.PresenterName != nil {
		trimmed := strings.TrimSpace(*parsed.PresenterName)
		if trimmed == "" {
			parsed.PresenterName = nil
		} else {
			parsed.PresenterName = &trimmed
		}
	}

	seenPlaces := make(map[string]struct{}, len(parsed.Spots))
	for i, item := range parsed.Spots {
		place := strings.TrimSpace(item.Place)
		if place == "" {
			return nil, fmt.Errorf("model output validation failed: spots[%d].place is required", i)
		}
		if item.SentenceStartTimestamp == nil {
			return nil, fmt.Errorf("model output validation failed: spots[%d].sentenceStartTimestamp is required", i)
		}
		if _, dup := seenPlaces[place]; dup {
			return nil, fmt.Errorf("model output validation failed: duplicate spots[%d].place %q found", i, place)
		}
		seenPlaces[place] = struct{}{}
		parsed.Spots[i].Place = place
		parsed.Spots[i].OriginalSentenceStartTimestamp = item.SentenceStartTimestamp
		parsed.Spots[i].RefinedSentenceStartTimestamp = item.SentenceStartTimestamp
	}

	return &parsed, nil
}

func ParseAndValidateRefinementResponse(rawBody []byte, pass1Spots []Candidate) (*ParsedRefinementResponse, error) {
	var parsed ParsedRefinementResponse
	if err := extractFunctionArgs(rawBody, SubmitRefinedSpotsFunctionName, &parsed); err != nil {
		return nil, err
	}

	allowed := make(map[string]struct{}, len(pass1Spots))
	for _, s := range pass1Spots {
		allowed[s.Place] = struct{}{}
	}
	seen := map[string]struct{}{}
	refinedByPlace := make(map[string]float64, len(parsed.Spots))
	for i, item := range parsed.Spots {
		place := strings.TrimSpace(item.Place)
		if place == "" {
			return nil, fmt.Errorf("model output validation failed: spots[%d].place is required", i)
		}
		if item.RefinedSentenceStartTimestamp == nil {
			return nil, fmt.Errorf("model output validation failed: spots[%d].refinedSentenceStartTimestamp is required", i)
		}
		if _, ok := allowed[place]; !ok {
			return nil, fmt.Errorf("model output validation failed: spots[%d].place=%q does not match pass-1 places", i, place)
		}
		if _, dup := seen[place]; dup {
			return nil, fmt.Errorf("model output validation failed: duplicate refinement for place %q", place)
		}
		seen[place] = struct{}{}
		parsed.Spots[i].Place = place
		refinedByPlace[place] = *item.RefinedSentenceStartTimestamp
	}

	var previousPass1 *float64
	for i, s := range pass1Spots {
		baseline := s.OriginalSentenceStartTimestamp
		if baseline == nil {
			baseline = s.SentenceStartTimestamp
		}
		if baseline == nil {
			return nil, fmt.Errorf("model output validation failed: pass-1 spots[%d].sentenceStartTimestamp is required for refinement validation", i)
		}

		if refined, ok := refinedByPlace[s.Place]; ok {
			if refined > *baseline {
				return nil, fmt.Errorf("model output validation failed: refinement for place %q must be <= pass-1 timestamp %.3f, got %.3f", s.Place, *baseline, refined)
			}
			if previousPass1 != nil && refined <= *previousPass1 {
				return nil, fmt.Errorf("model output validation failed: refinement for place %q must be > previous pass-1 timestamp %.3f, got %.3f", s.Place, *previousPass1, refined)
			}
		}

		v := *baseline
		previousPass1 = &v
	}

	return &parsed, nil
}

func ApplyRefinements(pass1 *ParsedResponse, refinement *ParsedRefinementResponse) *ParsedResponse {
	if pass1 == nil {
		return nil
	}

	refinedByPlace := map[string]float64{}
	if refinement != nil {
		for _, item := range refinement.Spots {
			if item.RefinedSentenceStartTimestamp == nil {
				continue
			}
			refinedByPlace[item.Place] = *item.RefinedSentenceStartTimestamp
		}
	}

	out := &ParsedResponse{PresenterName: pass1.PresenterName, Spots: make([]Candidate, len(pass1.Spots))}
	var previousPass1 *float64
	for i, s := range pass1.Spots {
		final := s.SentenceStartTimestamp
		if s.OriginalSentenceStartTimestamp != nil {
			final = s.OriginalSentenceStartTimestamp
		}
		baseline := final
		if original := s.OriginalSentenceStartTimestamp; original != nil {
			baseline = original
		}
		if baseline != nil {
			if refined, ok := refinedByPlace[s.Place]; ok {
				accept := refined <= *baseline
				if previousPass1 != nil {
					accept = accept && refined > *previousPass1
				}
				if accept {
					v := refined
					final = &v
				}
			}
		}

		var originalCopy *float64
		if s.OriginalSentenceStartTimestamp != nil {
			v := *s.OriginalSentenceStartTimestamp
			originalCopy = &v
		}
		var refinedCopy *float64
		if final != nil {
			v := *final
			refinedCopy = &v
		}
		out.Spots[i] = Candidate{
			Place:                          s.Place,
			SentenceStartTimestamp:         refinedCopy,
			OriginalSentenceStartTimestamp: originalCopy,
			RefinedSentenceStartTimestamp:  refinedCopy,
		}
		if baseline != nil {
			v := *baseline
			previousPass1 = &v
		}
	}

	return out
}

func extractFunctionArgs(rawBody []byte, functionName string, target any) error {
	var envelope generateContentEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("model response is not valid generateContent JSON: %w", err)
	}

	for _, cand := range envelope.Candidates {
		for _, part := range cand.Content.Parts {
			if part.FunctionCall == nil {
				continue
			}
			if part.FunctionCall.Name != functionName {
				continue
			}
			if len(part.FunctionCall.Args) == 0 {
				return fmt.Errorf("model function call args are empty")
			}
			if err := json.Unmarshal(part.FunctionCall.Args, target); err != nil {
				return fmt.Errorf("model function call args are not valid extraction JSON: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("model response does not contain extraction function call output")
}
