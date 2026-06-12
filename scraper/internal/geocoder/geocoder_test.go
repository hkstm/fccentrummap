package geocoder

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildTextSearchRequestBody_UsesRequiredRectangleWithoutLocationBias(t *testing.T) {
	body, err := buildTextSearchRequestBody("centrum")
	if err != nil {
		t.Fatalf("buildTextSearchRequestBody error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if _, ok := payload["locationBias"]; ok {
		t.Fatalf("locationBias must be absent")
	}
	if got := payload["languageCode"]; got != defaultLanguageCode {
		t.Fatalf("languageCode = %v, want %s", got, defaultLanguageCode)
	}

	rawRestriction, ok := payload["locationRestriction"].(map[string]any)
	if !ok {
		t.Fatalf("locationRestriction missing or invalid")
	}
	rawRect, ok := rawRestriction["rectangle"].(map[string]any)
	if !ok {
		t.Fatalf("rectangle missing or invalid")
	}
	low := rawRect["low"].(map[string]any)
	high := rawRect["high"].(map[string]any)

	assertFloat(t, low["latitude"], rectLowLat)
	assertFloat(t, low["longitude"], rectLowLng)
	assertFloat(t, high["latitude"], rectHighLat)
	assertFloat(t, high["longitude"], rectHighLng)
}

func TestGoogleMapsHTTPRefererDefaultsToProductionSite(t *testing.T) {
	t.Setenv("GOOGLE_MAPS_HTTP_REFERER", "")
	if got := googleMapsHTTPReferer(); got != defaultHTTPReferer {
		t.Fatalf("got %q, want %q", got, defaultHTTPReferer)
	}

	t.Setenv("GOOGLE_MAPS_HTTP_REFERER", "http://localhost:3000/")
	if got := googleMapsHTTPReferer(); got != "http://localhost:3000/" {
		t.Fatalf("got override %q", got)
	}
}

func TestParseCoordinatesFromTextSearchResponse_SelectsFirstValidResult(t *testing.T) {
	payload := []byte(`{"places":[{"location":{"latitude":0,"longitude":0}},{"id":"places/abc123","displayName":{"text":"Dream Unit"},"primaryType":"restaurant","primaryTypeDisplayName":{"text":"Restaurant"},"location":{"latitude":52.3,"longitude":4.9}},{"location":{"latitude":52.4,"longitude":4.8}}]}`)
	coords, err := parseCoordinatesFromTextSearchResponse(payload)
	if err != nil {
		t.Fatalf("parseCoordinatesFromTextSearchResponse error = %v", err)
	}
	if coords.Latitude != 52.3 || coords.Longitude != 4.9 {
		t.Fatalf("unexpected coordinates %+v", coords)
	}
	if coords.PlaceID != "places/abc123" {
		t.Fatalf("unexpected placeId: %q", coords.PlaceID)
	}
	if coords.Name != "Dream Unit" {
		t.Fatalf("unexpected name: %q", coords.Name)
	}
	if coords.PrimaryType != "restaurant" || coords.PrimaryTypeDisplayName != "Restaurant" {
		t.Fatalf("unexpected primary type fields: %+v", coords)
	}
}

func TestResolvePlaceInput(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/short":
			http.Redirect(w, r, srv.URL+"/maps/place/Van+Eeghenlaan+33-0,+1071+EP+Amsterdam/@52.3580636,4.8716319,17z", http.StatusFound)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/maps/place/"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/text":
			if got := r.Header.Get("X-Goog-FieldMask"); !strings.Contains(got, "places.primaryType") || !strings.Contains(got, "places.primaryTypeDisplayName") {
				t.Fatalf("missing primary type field mask: %q", got)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if !strings.Contains(string(body), "Van Eeghenlaan") {
				t.Fatalf("expected address text search body, got %s", string(body))
			}
			if !strings.Contains(string(body), `"languageCode":"nl"`) {
				t.Fatalf("expected Dutch languageCode, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"places":[{"id":"ChIJaddress","displayName":{"text":"Van Eeghenlaan 33"},"location":{"latitude":52.3580636,"longitude":4.8742068}}]}`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ChIJabc"):
			if got := r.URL.Query().Get("languageCode"); got != defaultLanguageCode {
				t.Fatalf("languageCode = %q, want %q", got, defaultLanguageCode)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ChIJabc","displayName":{"text":"Place"},"primaryType":"cafe","primaryTypeDisplayName":{"text":"Café"},"location":{"latitude":52.3,"longitude":4.9}}`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ChIJWTOQdOoJxkcRbp9D3FnPSoA"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ChIJWTOQdOoJxkcRbp9D3FnPSoA","displayName":{"text":"Ministry"},"location":{"latitude":52.3666158,"longitude":4.8896954}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	g := NewWithPlaceDetailsConfig("k", srv.URL+"/text", srv.URL+"/places", srv.Client())

	fromURL, err := g.ResolvePlaceInput(context.Background(), srv.URL+"/short")
	if err != nil {
		t.Fatalf("ResolvePlaceInput(short URL): %v", err)
	}
	if fromURL.PlaceID != "ChIJaddress" || fromURL.Latitude != 52.3580636 || fromURL.Longitude != 4.8742068 {
		t.Fatalf("unexpected maps URL coords %+v", fromURL)
	}

	fromFeatureTokenURL, err := g.ResolvePlaceInput(context.Background(), srv.URL+"/maps/place/Ministry/@52.3666158,4.8896954,17z/data=!3m1!4b1!4m6!3m5!1s0x47c609ea74903359:0x804acf59dc439f6e!8m2!3d52.3666158!4d4.8896954!16s%2Fg%2F11fmtm4td0")
	if err != nil {
		t.Fatalf("ResolvePlaceInput(feature token URL): %v", err)
	}
	if fromFeatureTokenURL.PlaceID != "ChIJWTOQdOoJxkcRbp9D3FnPSoA" || fromFeatureTokenURL.Latitude != 52.3666158 || fromFeatureTokenURL.Longitude != 4.8896954 {
		t.Fatalf("unexpected feature token URL coords %+v", fromFeatureTokenURL)
	}

	fromID, err := g.ResolvePlaceInput(context.Background(), "ChIJabc")
	if err != nil {
		t.Fatalf("ResolvePlaceInput(place ID): %v", err)
	}
	if fromID.PlaceID != "ChIJabc" || fromID.Latitude != 52.3 || fromID.Longitude != 4.9 {
		t.Fatalf("unexpected place ID coords %+v", fromID)
	}
	if fromID.PrimaryType != "cafe" || fromID.PrimaryTypeDisplayName != "Café" {
		t.Fatalf("unexpected place ID primary type fields %+v", fromID)
	}
}

func TestExtractPlaceInputFromMapsURL(t *testing.T) {
	if got, ok := extractPlaceQueryFromMapsURL("https://www.google.com/maps/place/Van+Eeghenlaan+33-0,+1071+EP+Amsterdam/@52.3580636,4.8716319,17z"); !ok || got != "Van Eeghenlaan 33-0, 1071 EP Amsterdam" {
		t.Fatalf("unexpected place query %q ok=%v", got, ok)
	}
	if got, ok := extractPlaceIDFromMapsURL("https://www.google.com/maps/search/?api=1&query=Place&query_place_id=ChIJabc"); !ok || got != "ChIJabc" {
		t.Fatalf("unexpected place ID %q ok=%v", got, ok)
	}
	if got, ok := extractPlaceIDFromMapsURL("https://www.google.com/maps?q=place_id:ChIJabc"); !ok || got != "ChIJabc" {
		t.Fatalf("unexpected q place ID %q ok=%v", got, ok)
	}
	if got, ok := extractPlaceIDFromMapsFeatureToken("https://www.google.com/maps/place/Ministry/@52.3666158,4.8896954,17z/data=!3m1!4b1!4m6!3m5!1s0x47c609ea74903359:0x804acf59dc439f6e!8m2!3d52.3666158!4d4.8896954"); !ok || got != "ChIJWTOQdOoJxkcRbp9D3FnPSoA" {
		t.Fatalf("unexpected feature token place ID %q ok=%v", got, ok)
	}
}

func TestLookupPlaceIDCoordinates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "ChIJabc") {
				t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
			}
			if got := r.Header.Get("X-Goog-FieldMask"); !strings.Contains(got, "primaryType") || !strings.Contains(got, "primaryTypeDisplayName") {
				t.Fatalf("expected primary type field mask header, got %q", got)
			}
			if got := r.URL.Query().Get("languageCode"); got != defaultLanguageCode {
				t.Fatalf("languageCode = %q, want %q", got, defaultLanguageCode)
			}
			if r.Header.Get("Referer") != defaultHTTPReferer {
				t.Fatalf("expected default referer header, got %q", r.Header.Get("Referer"))
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ChIJabc","displayName":{"text":"Place"},"primaryType":"cafe","primaryTypeDisplayName":{"text":"Café"},"location":{"latitude":52.3,"longitude":4.9}}`))
		}))
		defer srv.Close()

		g := NewWithPlaceDetailsConfig("k", "http://example.com/text", srv.URL, srv.Client())
		coords, err := g.LookupPlaceIDCoordinates(context.Background(), "ChIJabc")
		if err != nil {
			t.Fatalf("LookupPlaceIDCoordinates: %v", err)
		}
		if coords.PlaceID != "ChIJabc" || coords.Latitude != 52.3 || coords.Longitude != 4.9 {
			t.Fatalf("unexpected coords %+v", coords)
		}
		if coords.PrimaryType != "cafe" || coords.PrimaryTypeDisplayName != "Café" {
			t.Fatalf("unexpected primary type fields %+v", coords)
		}
	})

	t.Run("failure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer srv.Close()

		g := NewWithPlaceDetailsConfig("k", "http://example.com/text", srv.URL, srv.Client())
		_, err := g.LookupPlaceIDCoordinates(context.Background(), "bad")
		if err == nil || !contains(err.Error(), "HTTP 404") {
			t.Fatalf("expected HTTP 404, got %v", err)
		}
	})
}

func TestParseCoordinatesFromPlaceDetailsResponse_NoResults(t *testing.T) {
	_, err := parseCoordinatesFromPlaceDetailsResponse([]byte(`{"location":{"latitude":0,"longitude":0}}`))
	if !errors.Is(err, ErrNoResults) {
		t.Fatalf("expected ErrNoResults, got %v", err)
	}
}

func TestBuildStableMapsURL(t *testing.T) {
	got := BuildStableMapsURL("Dream Unit", "ChIJnybTiqkJxkcRVtXHKU6Lo-0")
	want := "https://www.google.com/maps/search/?api=1&query=Dream+Unit&query_place_id=ChIJnybTiqkJxkcRVtXHKU6Lo-0"
	if got != want {
		t.Fatalf("unexpected stable maps url: got %q want %q", got, want)
	}

	if BuildStableMapsURL("", "x") != "" {
		t.Fatalf("expected empty url when query is empty")
	}
	if BuildStableMapsURL("q", "") != "" {
		t.Fatalf("expected empty url when placeID is empty")
	}
}

func TestParseCoordinatesFromTextSearchResponse_NoResults(t *testing.T) {
	_, err := parseCoordinatesFromTextSearchResponse([]byte(`{"places":[]}`))
	if !errors.Is(err, ErrNoResults) {
		t.Fatalf("expected ErrNoResults, got %v", err)
	}
}

func TestParseCoordinatesFromTextSearchResponse_SkipsPartialCoordinates(t *testing.T) {
	payload := []byte(`{"places":[{"id":"places/partial","displayName":{"text":"Partial"},"location":{"latitude":52.3}},{"id":"places/full","displayName":{"text":"Full"},"location":{"latitude":52.31,"longitude":4.91}}]}`)
	coords, err := parseCoordinatesFromTextSearchResponse(payload)
	if err != nil {
		t.Fatalf("parseCoordinatesFromTextSearchResponse error = %v", err)
	}
	if coords.PlaceID != "places/full" || coords.Latitude != 52.31 || coords.Longitude != 4.91 {
		t.Fatalf("unexpected coordinates %+v", coords)
	}
}

func TestGeocodePlaceSendsRefererForBrowserRestrictedKey(t *testing.T) {
	t.Setenv("GOOGLE_MAPS_HTTP_REFERER", "https://fccentrummap.hkstm.dev/")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Referer"); got != "https://fccentrummap.hkstm.dev/" {
			t.Fatalf("Referer header = %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"places":[{"id":"ChIJabc","displayName":{"text":"Place"},"location":{"latitude":52.3,"longitude":4.9}}]}`))
	}))
	defer srv.Close()

	g := NewWithConfig("k", srv.URL, srv.Client())
	if _, err := g.GeocodePlace(context.Background(), "place"); err != nil {
		t.Fatalf("GeocodePlace: %v", err)
	}
}

func TestGeocodePlace_ErrorPaths(t *testing.T) {
	t.Run("missing API key", func(t *testing.T) {
		g := NewWithConfig("", "http://example.com", nil)
		_, err := g.GeocodePlace(context.Background(), "test")
		if !errors.Is(err, ErrMissingAPIKey) {
			t.Fatalf("expected ErrMissingAPIKey, got %v", err)
		}
	})

	t.Run("empty query", func(t *testing.T) {
		g := NewWithConfig("k", "http://example.com", nil)
		_, err := g.GeocodePlace(context.Background(), "   ")
		if !errors.Is(err, ErrEmptyQuery) {
			t.Fatalf("expected ErrEmptyQuery, got %v", err)
		}
	})

	t.Run("upstream http failure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusBadGateway)
		}))
		defer srv.Close()

		g := NewWithConfig("k", srv.URL, srv.Client())
		_, err := g.GeocodePlace(context.Background(), "centrum")
		if err == nil || !contains(err.Error(), "HTTP 502") {
			t.Fatalf("expected HTTP status error, got %v", err)
		}
	})

	t.Run("malformed response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not-json"))
		}))
		defer srv.Close()

		g := NewWithConfig("k", srv.URL, srv.Client())
		_, err := g.GeocodePlace(context.Background(), "centrum")
		if err == nil || !contains(err.Error(), "malformed places response") {
			t.Fatalf("expected malformed response error, got %v", err)
		}
	})
}

func assertFloat(t *testing.T, value any, want float64) {
	t.Helper()
	got, ok := value.(float64)
	if !ok {
		t.Fatalf("value %v is not float64", value)
	}
	if got != want {
		t.Fatalf("expected %v got %v", want, got)
	}
}

func contains(s, substr string) bool { return strings.Contains(s, substr) }
