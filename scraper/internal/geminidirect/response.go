package geminidirect

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Candidate struct {
	Place                   string   `json:"place"`
	Address                 *string  `json:"address"`
	YouTubeTimestampSeconds *float64 `json:"youtubeTimestampSeconds"`
	Evidence                string   `json:"evidence"`
	Confidence              *float64 `json:"confidence"`
	Notes                   string   `json:"notes"`
}

type ParsedResponse struct {
	ArticleURL    string      `json:"article_url,omitempty"`
	YouTubeURL    string      `json:"youtube_url,omitempty"`
	Model         string      `json:"model,omitempty"`
	PresenterName *string     `json:"presenter_name,omitempty"`
	Spots         []Candidate `json:"spots"`
}

type ParsedArtifact struct {
	ArticleSourceID int64       `json:"articleSourceId,omitempty"`
	ArticleURL      string      `json:"articleUrl"`
	YouTubeURL      string      `json:"youtubeUrl"`
	Model           string      `json:"model"`
	PresenterName   *string     `json:"presenterName,omitempty"`
	Spots           []Candidate `json:"spots"`
	Diagnostics     []string    `json:"diagnostics,omitempty"`
}

type generateContentEnvelope struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text    string `json:"text"`
				Thought bool   `json:"thought"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func ParseAndValidateResponse(rawBody []byte) (*ParsedResponse, error) {
	var parsed ParsedResponse
	if err := extractStructuredResponse(rawBody, &parsed); err != nil {
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
		if item.YouTubeTimestampSeconds != nil && *item.YouTubeTimestampSeconds < 0 {
			return nil, fmt.Errorf("model output validation failed: spots[%d].youtubeTimestampSeconds must be >= 0", i)
		}
		if item.Confidence != nil && (*item.Confidence < 0 || *item.Confidence > 1) {
			return nil, fmt.Errorf("model output validation failed: spots[%d].confidence must be between 0 and 1", i)
		}
		if item.Address != nil {
			address := strings.TrimSpace(*item.Address)
			if address == "" {
				return nil, fmt.Errorf("model output validation failed: spots[%d].address must be non-empty when provided", i)
			}
			parsed.Spots[i].Address = &address
		}
		evidence := strings.TrimSpace(item.Evidence)
		notes := strings.TrimSpace(item.Notes)
		if evidence == "" && notes == "" {
			return nil, fmt.Errorf("model output validation failed: spots[%d] requires evidence or notes", i)
		}
		key := strings.ToLower(place)
		if _, dup := seenPlaces[key]; dup {
			return nil, fmt.Errorf("model output validation failed: duplicate spots[%d].place %q found", i, place)
		}
		seenPlaces[key] = struct{}{}
		parsed.Spots[i].Place = place
		parsed.Spots[i].Evidence = evidence
		parsed.Spots[i].Notes = notes
	}

	return &parsed, nil
}

func ParseRawResponseToArtifact(rawBody []byte, articleSourceID int64, articleURL, youtubeURL, model string) (*ParsedArtifact, error) {
	parsed, err := ParseAndValidateResponse(rawBody)
	artifact := &ParsedArtifact{
		ArticleSourceID: articleSourceID,
		ArticleURL:      strings.TrimSpace(articleURL),
		YouTubeURL:      strings.TrimSpace(youtubeURL),
		Model:           strings.TrimSpace(model),
	}
	if err != nil {
		artifact.Diagnostics = []string{err.Error()}
		return artifact, err
	}
	if parsed.ArticleURL != "" && strings.TrimSpace(parsed.ArticleURL) != artifact.ArticleURL {
		err := fmt.Errorf("model output validation failed: article_url does not match requested article URL")
		artifact.Diagnostics = []string{err.Error()}
		return artifact, err
	}
	if parsed.YouTubeURL != "" && strings.TrimSpace(parsed.YouTubeURL) != artifact.YouTubeURL {
		err := fmt.Errorf("model output validation failed: youtube_url does not match requested YouTube URL")
		artifact.Diagnostics = []string{err.Error()}
		return artifact, err
	}
	if strings.TrimSpace(parsed.Model) != "" {
		artifact.Model = strings.TrimSpace(parsed.Model)
	}
	artifact.PresenterName = parsed.PresenterName
	artifact.Spots = parsed.Spots
	return artifact, nil
}

func extractStructuredResponse(rawBody []byte, target any) error {
	if err := unmarshalDirectStructuredJSON(rawBody, target); err == nil {
		return nil
	}

	var envelope generateContentEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("model response is not valid generateContent JSON: %w", err)
	}

	var lastErr error
	for _, cand := range envelope.Candidates {
		for _, part := range cand.Content.Parts {
			if part.Thought || strings.TrimSpace(part.Text) == "" {
				continue
			}
			if err := json.Unmarshal([]byte(strings.TrimSpace(part.Text)), target); err != nil {
				lastErr = err
				continue
			}
			return nil
		}
	}
	if lastErr != nil {
		return fmt.Errorf("model structured response text is not valid Gemini-direct JSON: %w", lastErr)
	}
	return fmt.Errorf("model response does not contain Gemini-direct structured JSON output")
}

func unmarshalDirectStructuredJSON(rawBody []byte, target any) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &fields); err != nil {
		return err
	}
	if _, ok := fields["spots"]; !ok {
		return fmt.Errorf("top-level spots field not found")
	}
	return json.Unmarshal(rawBody, target)
}
