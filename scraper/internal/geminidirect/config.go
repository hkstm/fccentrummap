package geminidirect

import gogenai "google.golang.org/genai"

func GenerateContentConfig() *gogenai.GenerateContentConfig {
	temperature := float32(0)
	return &gogenai.GenerateContentConfig{
		Temperature:      &temperature,
		Tools:            []*gogenai.Tool{{URLContext: &gogenai.URLContext{}}},
		ResponseMIMEType: "application/json",
		ResponseSchema:   responseSchema(),
		ThinkingConfig:   &gogenai.ThinkingConfig{IncludeThoughts: false},
	}
}

func responseSchema() *gogenai.Schema {
	return &gogenai.Schema{
		Type: gogenai.TypeObject,
		Properties: map[string]*gogenai.Schema{
			"presenter_name": {
				Type:        gogenai.TypeString,
				Nullable:    gogenai.Ptr(true),
				Description: "Optional primary presenter/person for the article/video.",
			},
			"spots": {
				Type:        gogenai.TypeArray,
				Description: "Timestamped spot candidates found from the URL context.",
				Items: &gogenai.Schema{
					Type: gogenai.TypeObject,
					Properties: map[string]*gogenai.Schema{
						"place": {
							Type: gogenai.TypeString,
							Description: "Canonical place name as it could be" +
								" listed on Google Maps.",
						},
						"address": {
							Type:        gogenai.TypeString,
							Nullable:    gogenai.Ptr(true),
							Description: "Optional address only when reasonably confident.",
						},
						"youtubeTimestampSeconds": {
							Type:        gogenai.TypeNumber,
							Description: "YouTube timestamp in seconds from the start of the video.",
						},
						"evidence": {
							Type:        gogenai.TypeString,
							Description: "Short source evidence supporting this spot.",
						},
						"confidence": {
							Type:        gogenai.TypeNumber,
							Description: "Confidence from 0.0 to 1.0.",
						},
						"notes": {
							Type:        gogenai.TypeString,
							Description: "Manual-review notes, ambiguity, or empty string when no note is needed.",
						},
					},
					Required: []string{"place", "youtubeTimestampSeconds", "evidence", "confidence", "notes"},
				},
			},
		},
		Required: []string{"spots"},
	}
}
