package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hkstm/fccentrummap/internal/geocoder"
)

func TestRun_SuccessJSON(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--query", "Dam Square"}, &out, &bytes.Buffer{}, func(ctx context.Context, placeName string) (*geocoder.Coordinates, error) {
		return &geocoder.Coordinates{Latitude: 52.3731, Longitude: 4.8922, PlaceID: "places/xyz", Name: "Dam Square"}, nil
	})
	if code != 0 {
		t.Fatalf("expected exit code 0 got %d", code)
	}

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if payload["query"] != "Dam Square" {
		t.Fatalf("expected query in output, got %v", payload["query"])
	}
	if payload["placeId"] != "places/xyz" {
		t.Fatalf("expected placeId in output, got %v", payload["placeId"])
	}
	if payload["name"] != "Dam Square" {
		t.Fatalf("expected name in output, got %v", payload["name"])
	}
	if payload["mapsUrl"] != "https://www.google.com/maps/search/?api=1&query=Dam+Square&query_place_id=places%2Fxyz" {
		t.Fatalf("expected mapsUrl in output, got %v", payload["mapsUrl"])
	}
	if payload["latitude"] != nil || payload["longitude"] != nil {
		t.Fatalf("did not expect latitude/longitude fields in output")
	}
}

func TestRun_ErrorJSONAndExitCode(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--query", "Dam"}, &out, &bytes.Buffer{}, func(ctx context.Context, placeName string) (*geocoder.Coordinates, error) {
		return nil, errors.New("upstream failure")
	})
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if payload["error"] == nil {
		t.Fatalf("expected error field in output")
	}
}
