package extractspots

type Request struct {
	DBPath     string
	OutDir     string
	Model string
	APIKey     string
	Endpoint   string
	InputPath  string
}

type Response struct {
	Identity          string  `json:"identity"`
	Stage             string  `json:"stage"`
	SpotExtractionIDs []int64 `json:"spotExtractionIds,omitempty"`
	OutputPath        string  `json:"outputPath,omitempty"`
}
