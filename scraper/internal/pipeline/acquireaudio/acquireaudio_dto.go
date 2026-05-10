package acquireaudio

type Request struct {
	DBPath    string
	Identity  string
	InputPath string
}

type AcquiredAudio struct {
	URL     string `json:"url"`
	VideoID string `json:"videoId"`
}

type Response struct {
	Identity   string          `json:"identity"`
	Stage      string          `json:"stage"`
	Acquired   []AcquiredAudio `json:"acquired,omitempty"`
	OutputPath string          `json:"outputPath,omitempty"`
}
