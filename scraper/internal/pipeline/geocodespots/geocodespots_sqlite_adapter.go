package geocodespots

import (
	"context"
	"fmt"
	"strings"

	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(ctx context.Context, req Request) (Response, error) {
	repo, err := repository.New(req.DBPath)
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}
	mentions, err := repo.ListSpotMentionsWithoutGeocode()
	if err != nil {
		return Response{}, err
	}
	g, err := geocoder.New()
	if err != nil {
		return Response{}, err
	}
	failures := make([]string, 0)
	for _, m := range mentions {
		coords, err := g.GeocodePlace(ctx, m.Place)
		if err != nil {
			failures = append(failures, fmt.Sprintf("spot_mention_id=%d place=%q: %v", m.SpotMentionID, m.Place, err))
			continue
		}
		status := "ok"
		formatted := m.Place
		placeID := coords.PlaceID
		if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpot(m.SpotMentionID, &placeID, coords.Latitude, coords.Longitude, &formatted, status, m.ArticleSourceID); err != nil {
			return Response{}, err
		}
	}
	if len(failures) > 0 {
		return Response{}, fmt.Errorf("geocode-spots completed with %d failures: %s", len(failures), strings.Join(failures, "; "))
	}
	return Response{Identity: "geocode-spots", Stage: "geocodespots"}, nil
}
