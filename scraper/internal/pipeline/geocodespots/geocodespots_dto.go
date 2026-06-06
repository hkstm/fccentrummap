package geocodespots

import "github.com/hkstm/fccentrummap/internal/models"

type Request struct {
	DBPath     string
	InputPath  string
	SpotSource models.SpotSource
}

type Response struct {
	Identity   string `json:"identity"`
	Stage      string `json:"stage"`
	OutputPath string `json:"outputPath,omitempty"`
}
