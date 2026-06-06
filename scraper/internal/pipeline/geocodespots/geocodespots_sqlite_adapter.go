package geocodespots

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/repository"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func normalizeAmsterdamQueryPart(value string) string {
	query := strings.TrimSpace(value)
	if query == "" {
		return ""
	}
	if !strings.Contains(strings.ToLower(query), "amsterdam") {
		query += ", Amsterdam"
	}
	return query
}

func (a *SQLiteAdapter) Run(ctx context.Context, req Request) (Response, error) {
	repo, err := repository.New(req.DBPath)
	if err != nil {
		return Response{}, err
	}
	defer repo.Close()
	if err := repo.InitSchema(); err != nil {
		return Response{}, err
	}
	mentions, err := repo.ListSpotMentionsWithoutGeocodeForSource(req.SpotSource)
	if err != nil {
		return Response{}, err
	}
	g, err := geocoder.New()
	if err != nil {
		return Response{}, err
	}
	failures := make([]string, 0)
	for _, m := range mentions {
		query := strings.TrimSpace(m.Place) + ", Amsterdam"
		fallbackQuery := ""
		if m.Address != nil && strings.TrimSpace(*m.Address) != "" {
			address := normalizeAmsterdamQueryPart(*m.Address)
			query = strings.TrimSpace(m.Place) + ", " + address
			fallbackQuery = address
		}
		coords, err := g.GeocodePlace(ctx, query)
		if err != nil && fallbackQuery != "" {
			coords, err = g.GeocodePlace(ctx, fallbackQuery)
		}
		if err != nil {
			failures = append(failures, fmt.Sprintf("spot_mention_id=%d place=%q: %v", m.SpotMentionID, m.Place, err))
			continue
		}
		status := "ok"
		formatted := m.Place
		placeID := coords.PlaceID
		primaryType := strings.TrimSpace(coords.PrimaryType)
		primaryTypeDisplayName := strings.TrimSpace(coords.PrimaryTypeDisplayName)
		metadata := repository.PlaceTypeMetadata{}
		if primaryType != "" {
			metadata.PrimaryType = &primaryType
		}
		if primaryTypeDisplayName != "" {
			metadata.PrimaryTypeDisplayName = &primaryTypeDisplayName
		}
		if _, err := repo.UpsertSpotGoogleGeocodeAndLinkArticleSpotForSource(req.SpotSource, m.SpotMentionID, &placeID, coords.Latitude, coords.Longitude, &formatted, status, m.ArticleSourceID, metadata); err != nil {
			return Response{}, err
		}
	}
	if len(failures) > 0 {
		fmt.Fprintf(os.Stderr, "geocode-spots completed with %d skipped failures: %s\n", len(failures), strings.Join(failures, "; "))
	}
	return Response{Identity: "geocode-spots", Stage: "geocodespots"}, nil
}
