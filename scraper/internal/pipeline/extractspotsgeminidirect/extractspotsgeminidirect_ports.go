package extractspotsgeminidirect

import (
	"context"

	genaiclient "github.com/hkstm/fccentrummap/internal/genai"
	gogenai "google.golang.org/genai"
)

type SQLitePort interface {
	Run(ctx context.Context, req Request) (Response, error)
}

type FilePort interface {
	Run(ctx context.Context, req Request) (Response, error)
}

type InputLoader interface {
	ListGeminiDirectInputs() ([]ArticleInput, error)
}

type ModelClient interface {
	GenerateContentWithParts(ctx context.Context, parts []*gogenai.Part, config *gogenai.GenerateContentConfig) (*genaiclient.GenerateContentResult, error)
}

type AcceptedSpot struct {
	ArticleSourceID         int64
	ArticleURL              string
	YouTubeURL              string
	Place                   string
	Address                 *string
	YouTubeTimestampSeconds *float64
	Evidence                *string
	Confidence              *float64
	Model                   *string
}

type AcceptedArticle struct {
	ArticleSourceID int64
	ArticleURL      string
	YouTubeURL      string
	PresenterName   *string
	Spots           []AcceptedSpot
}

type AcceptedSpotPersister interface {
	PersistAcceptedGeminiDirectArticle(article AcceptedArticle) error
}
