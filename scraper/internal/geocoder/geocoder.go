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
	defaultEndpoint             = "https://places.googleapis.com/v1/places:searchText"
	defaultPlaceDetailsEndpoint = "https://places.googleapis.com/v1/places"
	defaultHTTPReferer          = "https://fccentrummap.hkstm.dev/"

	rectLowLat  = 52.274525
	rectLowLng  = 4.711585
	rectHighLat = 52.461764
	rectHighLng = 5.073559
)

var (
	ErrMissingAPIKey = errors.New("missing PRODUCTION_GOOGLE_MAPS_API_KEY")
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
	apiKey               string
	endpoint             string
	placeDetailsEndpoint string
	httpClient           *http.Client
}

func New() (*Geocoder, error) {
	apiKey := strings.TrimSpace(os.Getenv("PRODUCTION_GOOGLE_MAPS_API_KEY"))
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	return NewWithPlaceDetailsConfig(
		apiKey,
		strings.TrimSpace(os.Getenv("GOOGLE_PLACES_TEXT_SEARCH_ENDPOINT")),
		strings.TrimSpace(os.Getenv("GOOGLE_PLACES_PLACE_DETAILS_ENDPOINT")),
		nil,
	), nil
}

func NewWithConfig(apiKey, endpoint string, httpClient *http.Client) *Geocoder {
	return NewWithPlaceDetailsConfig(apiKey, endpoint, "", httpClient)
}

func NewWithPlaceDetailsConfig(apiKey, endpoint, placeDetailsEndpoint string, httpClient *http.Client) *Geocoder {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	placeDetailsEndpoint = strings.TrimSpace(placeDetailsEndpoint)
	if placeDetailsEndpoint == "" {
		placeDetailsEndpoint = defaultPlaceDetailsEndpoint
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Geocoder{apiKey: strings.TrimSpace(apiKey), endpoint: endpoint, placeDetailsEndpoint: placeDetailsEndpoint, httpClient: httpClient}
}

func (g *Geocoder) setGoogleAPIHeaders(req *http.Request, fieldMask string) {
	req.Header.Set("X-Goog-Api-Key", g.apiKey)
	req.Header.Set("X-Goog-FieldMask", fieldMask)
	if referer := googleMapsHTTPReferer(); referer != "" {
		req.Header.Set("Referer", referer)
	}
}

func googleMapsHTTPReferer() string {
	if v := strings.TrimSpace(os.Getenv("GOOGLE_MAPS_HTTP_REFERER")); v != "" {
		return v
	}
	return defaultHTTPReferer
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
	g.setGoogleAPIHeaders(req, "places.location,places.id,places.displayName.text")

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

func (g *Geocoder) ResolvePlaceInput(ctx context.Context, input string) (*Coordinates, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return nil, ErrEmptyQuery
	}

	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		if placeID, ok := extractPlaceIDFromMapsURL(value); ok {
			return g.LookupPlaceIDCoordinates(ctx, placeID)
		}
		if query, ok := extractPlaceQueryFromMapsURL(value); ok {
			return g.GeocodePlace(ctx, query)
		}

		resolved := value
		if finalURL, err := g.resolveRedirectURL(ctx, value); err == nil {
			resolved = finalURL
		}
		if placeID, ok := extractPlaceIDFromMapsURL(resolved); ok {
			return g.LookupPlaceIDCoordinates(ctx, placeID)
		}
		if query, ok := extractPlaceQueryFromMapsURL(resolved); ok {
			return g.GeocodePlace(ctx, query)
		}
		return nil, fmt.Errorf("could not extract place ID or address from maps URL %q", value)
	}

	if placeID, ok := strings.CutPrefix(value, "place_id:"); ok {
		return g.LookupPlaceIDCoordinates(ctx, placeID)
	}
	if isLikelyGooglePlaceID(value) {
		return g.LookupPlaceIDCoordinates(ctx, value)
	}
	return g.GeocodePlace(ctx, value)
}

func (g *Geocoder) resolveRedirectURL(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

func isLikelyGooglePlaceID(value string) bool {
	v := strings.TrimSpace(value)
	return strings.HasPrefix(v, "ChIJ") || strings.HasPrefix(v, "places/")
}

func extractPlaceIDFromMapsURL(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	params := u.Query()
	for _, key := range []string{"query_place_id", "place_id"} {
		if id := strings.TrimSpace(params.Get(key)); id != "" {
			return id, true
		}
	}
	if q := strings.TrimSpace(params.Get("q")); strings.HasPrefix(q, "place_id:") {
		return strings.TrimSpace(strings.TrimPrefix(q, "place_id:")), true
	}
	return "", false
}

func extractPlaceQueryFromMapsURL(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	params := u.Query()
	for _, key := range []string{"query", "q"} {
		if q := strings.TrimSpace(params.Get(key)); q != "" && !strings.HasPrefix(q, "place_id:") {
			return q, true
		}
	}

	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	for i, part := range parts {
		if part != "place" || i+1 >= len(parts) {
			continue
		}
		query, err := url.PathUnescape(parts[i+1])
		if err != nil {
			return "", false
		}
		query = strings.TrimSpace(strings.ReplaceAll(query, "+", " "))
		if query != "" {
			return query, true
		}
	}
	return "", false
}

func (g *Geocoder) LookupPlaceIDCoordinates(ctx context.Context, placeID string) (*Coordinates, error) {
	if strings.TrimSpace(g.apiKey) == "" {
		return nil, ErrMissingAPIKey
	}
	id := strings.TrimSpace(placeID)
	if id == "" {
		return nil, ErrEmptyQuery
	}

	endpoint := strings.TrimRight(g.placeDetailsEndpoint, "/") + "/" + url.PathEscape(id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create place details request: %w", err)
	}
	g.setGoogleAPIHeaders(req, "id,location,displayName.text")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("place-id lookup request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read place-id lookup response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("place-id lookup HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	coords, err := parseCoordinatesFromPlaceDetailsResponse(respBody)
	if err != nil {
		return nil, err
	}
	if coords.PlaceID == "" {
		coords.PlaceID = id
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
		ID          string `json:"id"`
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

type placeDetailsResponse struct {
	ID          string `json:"id"`
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	Location struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	} `json:"location"`
	Error any `json:"error"`
}

func parseCoordinatesFromPlaceDetailsResponse(body []byte) (*Coordinates, error) {
	var payload placeDetailsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("malformed place details response: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("place details API returned error: %v", payload.Error)
	}
	if payload.Location.Latitude == nil || payload.Location.Longitude == nil {
		return nil, ErrNoResults
	}
	lat := *payload.Location.Latitude
	lng := *payload.Location.Longitude
	if lat == 0 && lng == 0 {
		return nil, ErrNoResults
	}
	return &Coordinates{
		Latitude:  lat,
		Longitude: lng,
		PlaceID:   strings.TrimSpace(payload.ID),
		Name:      strings.TrimSpace(payload.DisplayName.Text),
	}, nil
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
