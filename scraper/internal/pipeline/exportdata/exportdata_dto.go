package exportdata

type Request struct {
	DBPath     string
	OutputPath string
	Identity   string
	InputPath  string
}

type Response struct {
	Identity   string `json:"identity"`
	Stage      string `json:"stage"`
	OutputPath string `json:"outputPath,omitempty"`
}
