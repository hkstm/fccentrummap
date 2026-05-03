package geocoder

import (
	"context"
	"encoding/json"
	"errors"
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

func TestParseCoordinatesFromTextSearchResponse_SelectsFirstValidResult(t *testing.T) {
	payload := []byte(`{"places":[{"location":{"latitude":0,"longitude":0}},{"id":"places/abc123","displayName":{"text":"Dream Unit"},"location":{"latitude":52.3,"longitude":4.9}},{"location":{"latitude":52.4,"longitude":4.8}}]}`)
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
