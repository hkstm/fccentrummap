package extractarticletext

type Request struct {
	DBPath    string
	InputPath string
}

type Response struct {
	Identity       string `json:"identity"`
	Stage          string `json:"stage"`
	ProcessedCount int    `json:"processedCount,omitempty"`
	OutputPath     string `json:"outputPath,omitempty"`
}
