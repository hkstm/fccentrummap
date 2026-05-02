package extraction

import gogenai "google.golang.org/genai"

const (
	SubmitSpotsFunctionName        = "submit_spots"
	SubmitRefinedSpotsFunctionName = "submit_refined_spots"
)

func GenerateContentConfig() *gogenai.GenerateContentConfig {
	return GeneratePass1ContentConfig()
}

func GeneratePass1ContentConfig() *gogenai.GenerateContentConfig {
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
		ToolConfig:     functionCallingConfig(SubmitSpotsFunctionName),
		ThinkingConfig: &gogenai.ThinkingConfig{IncludeThoughts: false},
	}
}

func GeneratePass2RefinementContentConfig() *gogenai.GenerateContentConfig {
	temperature := float32(0)
	return &gogenai.GenerateContentConfig{
		Temperature: &temperature,
		Tools: []*gogenai.Tool{{
			FunctionDeclarations: []*gogenai.FunctionDeclaration{{
				Name:        SubmitRefinedSpotsFunctionName,
				Description: "Return refined timestamp anchors for pass-1 spots. Each refined timestamp must be earlier-or-equal to the original timestamp.",
				Parameters: &gogenai.Schema{
					Type: gogenai.TypeObject,
					Properties: map[string]*gogenai.Schema{
						"spots": {
							Type: gogenai.TypeArray,
							Items: &gogenai.Schema{
								Type: gogenai.TypeObject,
								Properties: map[string]*gogenai.Schema{
									"place": {
										Type:        gogenai.TypeString,
										Description: "Exact place label from pass 1.",
									},
									"refinedSentenceStartTimestamp": {
										Type:        gogenai.TypeNumber,
										Description: "Refined sentence start timestamp for this place.",
									},
								},
								Required: []string{"place", "refinedSentenceStartTimestamp"},
							},
						},
					},
					Required: []string{"spots"},
				},
			}},
		}},
		ToolConfig:     functionCallingConfig(SubmitRefinedSpotsFunctionName),
		ThinkingConfig: &gogenai.ThinkingConfig{IncludeThoughts: false},
	}
}

func functionCallingConfig(fn string) *gogenai.ToolConfig {
	return &gogenai.ToolConfig{
		FunctionCallingConfig: &gogenai.FunctionCallingConfig{
			Mode:                 gogenai.FunctionCallingConfigModeAny,
			AllowedFunctionNames: []string{fn},
		},
	}
}
