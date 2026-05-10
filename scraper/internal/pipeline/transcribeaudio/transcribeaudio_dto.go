package transcribeaudio

type Request struct {
	DBPath    string
	Language  string
	Identity  string
	InputPath string
}

type Response struct {
	Identity        string `json:"identity"`
	Stage           string `json:"stage"`
	TranscriptionID int64  `json:"transcriptionId,omitempty"`
	OutputPath      string `json:"outputPath,omitempty"`
}
