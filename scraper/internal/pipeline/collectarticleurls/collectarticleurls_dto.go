package collectarticleurls

type Request struct {
	DBPath     string
	ArticleURL string
	InputPath  string
}

type Response struct {
	Identity   string   `json:"identity"`
	Stage      string   `json:"stage"`
	URLs       []string `json:"articleUrls,omitempty"`
	OutputPath string   `json:"outputPath,omitempty"`
}
