package extractspots

type Request struct {
	DBPath     string
	OutDir     string
	GemmaModel string
	APIKey     string
	Endpoint   string
	InputPath  string
}

type Response struct {
	Identity         string `json:"identity"`
	Stage            string `json:"stage"`
	SpotExtractionID int64  `json:"spotExtractionId,omitempty"`
	OutputPath       string `json:"outputPath,omitempty"`
}
