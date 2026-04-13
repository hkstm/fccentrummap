package geocoder

import (
	"context"
	"fmt"
	"log"
	"os"

	"googlemaps.github.io/maps"
)

type Geocoder struct {
	client *maps.Client
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func New() (*Geocoder, error) {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Printf("ERROR: GOOGLE_MAPS_API_KEY environment variable is not set")
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY environment variable is not set")
	}

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("creating maps client: %w", err)
	}

	return &Geocoder{client: client}, nil
}

func (g *Geocoder) Geocode(name, address string) (*Coordinates, error) {
	query := address + ", Amsterdam"

	results, err := g.client.Geocode(context.Background(), &maps.GeocodingRequest{
		Address: query,
	})
	if err != nil {
		log.Printf("ERROR: geocoding failed spot=%q address=%q error=%v", name, address, err)
		return nil, fmt.Errorf("geocoding spot=%q address=%q: %w", name, address, err)
	}

	if len(results) == 0 {
		log.Printf("ERROR: geocoding returned no results spot=%q address=%q", name, address)
		return nil, fmt.Errorf("geocoding returned no results for spot=%q address=%q", name, address)
	}

	loc := results[0].Geometry.Location
	return &Coordinates{
		Latitude:  loc.Lat,
		Longitude: loc.Lng,
	}, nil
}
