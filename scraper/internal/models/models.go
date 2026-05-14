package models

import "time"

type ArticleSource struct {
	ArticleSourceID int64
	URL             string
	DiscoveredAt    time.Time
}

type ArticleFetch struct {
	ArticleFetchID  int64
	ArticleSourceID int64
	HTML            string
	FetchedAt       time.Time
}

type ArticleText struct {
	ArticleTextID   int64
	ArticleFetchID  int64
	CleanedText     string
	ExtractedAt     time.Time
}

type ArticleRaw struct {
	ArticleRawID int64
	URL          string
	HTML         string
	VideoID      *string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ArticleAudioSource struct {
	AudioSourceID int64
	ArticleRawID  int64
	VideoID       string
	YouTubeURL    string
	AudioFormat   string
	MIMEType      string
	AudioBlob     []byte
	ByteSize      int64
	CreatedAt     time.Time
}

type ArticleAudioTranscription struct {
	TranscriptionID  int64
	AudioSourceID    int64
	Provider         string
	Language         string
	HTTPStatus       int
	ResponseJSON     string
	ResponseByteSize int64
	ErrorMessage     *string
	CreatedAt        time.Time
}

const (
	ArticleTextExtractionStatusMatched = "matched"
	ArticleTextExtractionStatusNoMatch = "no_match"
	ArticleTextExtractionStatusError   = "error"

	ArticleTextExtractionModeTrafilatura = "trafilatura"
	ArticleTextExtractionModeNoMatch     = "no_match"
	ArticleTextExtractionModeError       = "error"

	ArticleTextSourceTypeTrafilaturaText = "trafilatura-text"
)

type ArticleTextContentInput struct {
	SourceType string
	Content    string
}

type ArticleTextExtractionResult struct {
	ArticleRawID   int64
	ExtractionMode string
	Status         string
	MatchedCount   int
	ErrorMessage   *string
	Contents       []ArticleTextContentInput
}

type ArticleTextExtraction struct {
	ExtractionID   int64
	ArticleRawID   int64
	ExtractionMode string
	Status         string
	MatchedCount   int
	ErrorMessage   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ArticleTextContent struct {
	TextContentID int64
	ExtractionID  int64
	ArticleRawID  int64
	SourceType    string
	Content       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Author struct {
	AuthorID int64
	Name     string
}

type Spot struct {
	SpotID    int64
	Name      string
	Address   string
	Latitude  float64
	Longitude float64
}

type Article struct {
	ArticleID    int64
	ArticleRawID int64
	AuthorID     int64
	Title        string
}

type ArticleSpot struct {
	ArticleID int64
	SpotID    int64
}

type ExportSpot struct {
	PlaceID       string  `json:"placeId"`
	SpotName      string  `json:"spotName"`
	PresenterName string  `json:"presenterName"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	YouTubeLink   string  `json:"youtubeLink"`
	ArticleURL    string  `json:"articleUrl"`
}

type ExportPresenter struct {
	PresenterName string `json:"presenterName"`
}

type SpotExtractionRecordInput struct {
	ArticleRawID       int64
	TranscriptionID    int64
	PresenterName      *string
	PromptText         string
	RawResponseJSON    string
	ParsedResponseJSON string
}

type SpotExtractionRecord struct {
	SpotExtractionID   int64
	ArticleRawID       int64
	TranscriptionID    int64
	PresenterName      *string
	PromptText         string
	RawResponseJSON    string
	ParsedResponseJSON string
	CreatedAt          time.Time
}

type ExportData struct {
	Spots      []ExportSpot      `json:"spots"`
	Presenters []ExportPresenter `json:"presenters"`
}
