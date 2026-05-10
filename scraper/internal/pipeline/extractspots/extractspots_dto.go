package extractspots

type Request struct {
	DBPath      string
	OutDir      string
	GemmaModel  string
	APIKey      string
	Endpoint    string
	Identity    string
	InputPath   string
	UseLatest   bool
	PersistData bool
}

type Response struct {
	Identity         string `json:"identity"`
	Stage            string `json:"stage"`
	SpotExtractionID int64  `json:"spotExtractionId,omitempty"`
	OutputPath       string `json:"outputPath,omitempty"`
}
