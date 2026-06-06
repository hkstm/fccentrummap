package extractspotsgeminidirect

type Request struct {
	DBPath   string
	OutDir   string
	Model    string
	APIKey   string
	Endpoint string
	Force    bool
	Limit    int
}

type Response struct {
	Identity       string          `json:"identity"`
	Stage          string          `json:"stage"`
	ProcessedCount int             `json:"processedCount"`
	SkippedCount   int             `json:"skippedCount"`
	CachedCount    int             `json:"cachedCount"`
	OutputDir      string          `json:"outputDir"`
	Articles       []ArticleResult `json:"articles,omitempty"`
}

type ArticleInput struct {
	ArticleSourceID int64
	ArticleURL      string
	YouTubeURL      string
}

type ArticleResult struct {
	ArticleSourceID int64    `json:"articleSourceId"`
	ArticleURL      string   `json:"articleUrl"`
	YouTubeURL      string   `json:"youtubeUrl,omitempty"`
	PromptPath      string   `json:"promptPath,omitempty"`
	RawResponsePath string   `json:"rawResponsePath,omitempty"`
	ParsedPath      string   `json:"parsedPath,omitempty"`
	Diagnostics     []string `json:"diagnostics,omitempty"`
	Skipped         bool     `json:"skipped,omitempty"`
	Cached          bool     `json:"cached,omitempty"`
}
