package models

import (
	"fmt"
	"strings"
	"time"
)

// SpotSource selects which extracted spot table set downstream stages consume.
type SpotSource string

const (
	SpotSourceGeminiDirect SpotSource = "gemini-direct"
)

func NormalizeSpotSource(raw string) (SpotSource, error) {
	source := SpotSource(strings.TrimSpace(strings.ToLower(raw)))
	if source == "" {
		return SpotSourceGeminiDirect, nil
	}
	switch source {
	case SpotSourceGeminiDirect:
		return source, nil
	default:
		return "", fmt.Errorf("unsupported spot source %q (supported: %s)", raw, SpotSourceGeminiDirect)
	}
}

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

type ArticleRaw struct {
	ArticleRawID int64
	URL          string
	HTML         string
	VideoID      *string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	SpotID        string  `json:"spotId"`
	PlaceID       string  `json:"placeId"`
	SpotName      string  `json:"spotName"`
	PresenterName string  `json:"presenterName"`
	CategoryName  string  `json:"categoryName"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	YouTubeLink   string  `json:"youtubeLink"`
	ArticleURL    string  `json:"articleUrl"`
}

type ExportPresenter struct {
	PresenterName string `json:"presenterName"`
}

type ExportCategory struct {
	CategoryName string `json:"categoryName"`
}

// SpotCorrection stores optional maintainer overrides keyed by stable ExportSpot.SpotID.
type SpotCorrection struct {
	SpotID                  string
	SpotName                *string
	PlaceID                 *string
	Latitude                *float64
	Longitude               *float64
	YouTubeTimestampSeconds *int64
	Hidden                  bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// SpotCorrectionTarget exposes source and currently-effective values for correction prompts.
type SpotCorrectionTarget struct {
	SpotID                        string
	SourceSpotName                string
	SourcePlaceID                 string
	SourceLatitude                float64
	SourceLongitude               float64
	SourceYouTubeLink             string
	SourceYouTubeTimestampSecs    *int64
	EffectiveSpotName             string
	EffectivePlaceID              string
	EffectiveLatitude             float64
	EffectiveLongitude            float64
	EffectiveYouTubeLink          string
	EffectiveYouTubeTimestampSecs *int64
	ArticleURL                    string
	PresenterName                 string
	Correction                    *SpotCorrection
}

type ExportData struct {
	Presenters []ExportPresenter `json:"presenters,omitempty"`
	Categories []ExportCategory  `json:"categories,omitempty"`
	Spots      []ExportSpot      `json:"spots"`
}
