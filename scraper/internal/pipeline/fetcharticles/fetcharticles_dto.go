package fetcharticles

type Request struct {
	DBPath    string
	InputPath string
}

type Response struct {
	Identity     string   `json:"identity"`
	Stage        string   `json:"stage"`
	ArticleURLs  []string `json:"articleUrls,omitempty"`
	FetchedCount int      `json:"fetchedCount,omitempty"`
	OutputPath   string   `json:"outputPath,omitempty"`
}
