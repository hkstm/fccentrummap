package geocodespots

type Request struct {
	DBPath    string
	InputPath string
}

type Response struct {
	Identity   string `json:"identity"`
	Stage      string `json:"stage"`
	OutputPath string `json:"outputPath,omitempty"`
}
