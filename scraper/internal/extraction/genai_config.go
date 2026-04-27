package extraction

import gogenai "google.golang.org/genai"

const SubmitSpotsFunctionName = "submit_spots"

func GenerateContentConfig() *gogenai.GenerateContentConfig {
	temperature := float32(0)
	return &gogenai.GenerateContentConfig{
		Temperature: &temperature,
		Tools: []*gogenai.Tool{{
			FunctionDeclarations: []*gogenai.FunctionDeclaration{{
				Name:        SubmitSpotsFunctionName,
				Description: "Return extracted Amsterdam-area spots using cleaned article as primary source and transcript evidence for timestamps.",
				Parameters: &gogenai.Schema{
					Type: gogenai.TypeObject,
					Properties: map[string]*gogenai.Schema{
						"presenter_name": {
							Type:        gogenai.TypeString,
							Nullable:    gogenai.Ptr(true),
							Description: "Optional primary presenter/person for the extraction run.",
						},
						"spots": {
							Type: gogenai.TypeArray,
							Items: &gogenai.Schema{
								Type: gogenai.TypeObject,
								Properties: map[string]*gogenai.Schema{
									"place": {
										Type:        gogenai.TypeString,
										Description: "Name of place mentioned in transcript.",
									},
									"sentenceStartTimestamp": {
										Type:        gogenai.TypeNumber,
										Description: "Sentence start timestamp from evidence.",
									},
								},
								Required: []string{"place", "sentenceStartTimestamp"},
							},
						},
					},
					Required: []string{"spots"},
				},
			}},
		}},
		ToolConfig: &gogenai.ToolConfig{
			FunctionCallingConfig: &gogenai.FunctionCallingConfig{
				Mode:                 gogenai.FunctionCallingConfigModeAny,
				AllowedFunctionNames: []string{SubmitSpotsFunctionName},
			},
		},
		ThinkingConfig: &gogenai.ThinkingConfig{IncludeThoughts: false},
	}
}
