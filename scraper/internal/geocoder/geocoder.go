package geocoder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultEndpoint = "https://places.googleapis.com/v1/places:searchText"

	rectLowLat  = 52.274525
	rectLowLng  = 4.711585
	rectHighLat = 52.461764
	rectHighLng = 5.073559
)

var (
	ErrMissingAPIKey = errors.New("missing GOOGLE_MAPS_API_KEY")
	ErrEmptyQuery    = errors.New("place query is required")
	ErrNoResults     = errors.New("no place match found within enforced location restriction")
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	PlaceID   string  `json:"placeId,omitempty"`
	Name      string  `json:"name,omitempty"`
}

type Geocoder struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

func New() (*Geocoder, error) {
	apiKey := strings.TrimSpace(os.Getenv("GOOGLE_MAPS_API_KEY"))
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	return NewWithConfig(apiKey, strings.TrimSpace(os.Getenv("GOOGLE_PLACES_TEXT_SEARCH_ENDPOINT")), nil), nil
}

func NewWithConfig(apiKey, endpoint string, httpClient *http.Client) *Geocoder {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Geocoder{apiKey: strings.TrimSpace(apiKey), endpoint: endpoint, httpClient: httpClient}
}

func (g *Geocoder) GeocodePlace(ctx context.Context, placeName string) (*Coordinates, error) {
	if strings.TrimSpace(g.apiKey) == "" {
		return nil, ErrMissingAPIKey
	}
	query := strings.TrimSpace(placeName)
	if query == "" {
		return nil, ErrEmptyQuery
	}

	body, err := buildTextSearchRequestBody(query)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", g.apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.location,places.id,places.displayName.text")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("places text search request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read places response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("places text search HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	coords, err := parseCoordinatesFromTextSearchResponse(respBody)
	if err != nil {
		return nil, err
	}
	return coords, nil
}

func buildTextSearchRequestBody(query string) ([]byte, error) {
	payload := map[string]any{
		"textQuery": query,
		"locationRestriction": map[string]any{
			"rectangle": map[string]any{
				"low": map[string]float64{
					"latitude":  rectLowLat,
					"longitude": rectLowLng,
				},
				"high": map[string]float64{
					"latitude":  rectHighLat,
					"longitude": rectHighLng,
				},
			},
		},
	}
	return json.Marshal(payload)
}

type textSearchResponse struct {
	Places []struct {
		ID string `json:"id"`
		DisplayName struct {
			Text string `json:"text"`
		} `json:"displayName"`
		Location struct {
			Latitude  *float64 `json:"latitude"`
			Longitude *float64 `json:"longitude"`
		} `json:"location"`
	} `json:"places"`
	Error any `json:"error"`
}

func parseCoordinatesFromTextSearchResponse(body []byte) (*Coordinates, error) {
	var payload textSearchResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("malformed places response: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("places API returned error: %v", payload.Error)
	}
	for _, p := range payload.Places {
		if p.Location.Latitude == nil || p.Location.Longitude == nil {
			continue
		}
		lat := *p.Location.Latitude
		lng := *p.Location.Longitude
		if lat == 0 && lng == 0 {
			continue
		}
		return &Coordinates{
			Latitude:  lat,
			Longitude: lng,
			PlaceID:   strings.TrimSpace(p.ID),
			Name:      strings.TrimSpace(p.DisplayName.Text),
		}, nil
	}
	return nil, ErrNoResults
}

func BuildStableMapsURL(query, placeID string) string {
	q := strings.TrimSpace(query)
	id := strings.TrimSpace(placeID)
	if q == "" || id == "" {
		return ""
	}
	return "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(q) + "&query_place_id=" + url.QueryEscape(id)
}
