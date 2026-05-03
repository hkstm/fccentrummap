package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hkstm/fccentrummap/internal/geocoder"
)

type geocodeFunc func(ctx context.Context, placeName string) (*geocoder.Coordinates, error)

func main() {
	if code := run(os.Args[1:], os.Stdout, os.Stderr, defaultGeocodeFunc()); code != 0 {
		os.Exit(code)
	}
}

func run(args []string, stdout, stderr io.Writer, geocode geocodeFunc) int {
	fs := flag.NewFlagSet("geocode-place", flag.ContinueOnError)
	fs.SetOutput(stderr)
	query := fs.String("query", "", "place name query")
	if err := fs.Parse(args); err != nil {
		return printJSONError(stdout, fmt.Errorf("invalid flags: %w", err), 2)
	}
	if strings.TrimSpace(*query) == "" && fs.NArg() > 0 {
		*query = strings.Join(fs.Args(), " ")
	}
	if strings.TrimSpace(*query) == "" {
		return printJSONError(stdout, geocoder.ErrEmptyQuery, 2)
	}

	coords, err := geocode(context.Background(), *query)
	if err != nil {
		code := 1
		if errors.Is(err, geocoder.ErrEmptyQuery) || errors.Is(err, geocoder.ErrMissingAPIKey) {
			code = 2
		}
		return printJSONError(stdout, err, code)
	}

	trimmedQuery := strings.TrimSpace(*query)
	resp := map[string]any{
		"query": trimmedQuery,
	}
	if coords.PlaceID != "" {
		resp["placeId"] = coords.PlaceID
	}
	if coords.Name != "" {
		resp["name"] = coords.Name
	}
	if stableURL := geocoder.BuildStableMapsURL(trimmedQuery, coords.PlaceID); stableURL != "" {
		resp["mapsUrl"] = stableURL
	}
	enc := json.NewEncoder(stdout)
	if err := enc.Encode(resp); err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to encode json output: %v\n", err)
		return 1
	}
	return 0
}

func defaultGeocodeFunc() geocodeFunc {
	return func(ctx context.Context, placeName string) (*geocoder.Coordinates, error) {
		g, err := geocoder.New()
		if err != nil {
			return nil, err
		}
		return g.GeocodePlace(ctx, placeName)
	}
}

func printJSONError(stdout io.Writer, err error, code int) int {
	_ = json.NewEncoder(stdout).Encode(map[string]any{
		"error": err.Error(),
	})
	return code
}
