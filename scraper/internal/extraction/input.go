package extraction

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SentenceUnit struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
}

type transcriptionPayload struct {
	Text     string                 `json:"text"`
	Segments []transcriptionSegment `json:"segments"`
	Words    []any                  `json:"words"`
}

type transcriptionSegment struct {
	Text  string   `json:"text"`
	Start *float64 `json:"start"`
}

func ParseSentenceUnits(responseJSON string) ([]SentenceUnit, error) {
	var payload transcriptionPayload
	if err := json.Unmarshal([]byte(responseJSON), &payload); err != nil {
		return nil, fmt.Errorf("transcript payload is not valid JSON: %w", err)
	}

	if len(payload.Segments) == 0 {
		if strings.TrimSpace(payload.Text) != "" {
			return nil, fmt.Errorf("transcript does not contain sentence-level segments with timestamps; full-text-only payloads are not supported")
		}
		if len(payload.Words) > 0 {
			return nil, fmt.Errorf("transcript does not contain sentence-level segments with timestamps; word-level-only payloads are not supported")
		}
		return nil, fmt.Errorf("transcript does not contain sentence-level segments with timestamps")
	}

	sentences := make([]SentenceUnit, 0, len(payload.Segments))
	for i, seg := range payload.Segments {
		text := strings.TrimSpace(seg.Text)
		if text == "" {
			continue
		}
		if seg.Start == nil {
			return nil, fmt.Errorf("segment %d is missing required start timestamp", i)
		}
		sentences = append(sentences, SentenceUnit{Text: text, Start: *seg.Start})
	}

	if len(sentences) == 0 {
		return nil, fmt.Errorf("transcript does not contain usable sentence-level segments with both text and start timestamp")
	}

	return sentences, nil
}
