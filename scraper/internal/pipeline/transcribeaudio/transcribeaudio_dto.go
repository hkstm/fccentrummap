package transcribeaudio

type Request struct {
	DBPath    string
	Language  string
	InputPath string
}

type Response struct {
	Identity         string  `json:"identity"`
	Stage            string  `json:"stage"`
	TranscriptionIDs []int64 `json:"transcriptionIds,omitempty"`
	OutputPath       string  `json:"outputPath,omitempty"`
}
