package models

import "time"

type ArticleRaw struct {
	ArticleRawID int64
	URL          string
	HTML         string
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
	Name    string   `json:"name"`
	Address string   `json:"address"`
	Lat     float64  `json:"lat"`
	Lng     float64  `json:"lng"`
	Authors []string `json:"authors"`
}

type ExportData struct {
	Authors []string     `json:"authors"`
	Spots   []ExportSpot `json:"spots"`
}
