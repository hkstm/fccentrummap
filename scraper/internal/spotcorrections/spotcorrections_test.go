package spotcorrections

import (
	"context"
	"errors"
	"testing"

	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/models"
)

type fakeRepo struct {
	target *models.SpotCorrectionTarget
	saved  *models.SpotCorrection
}

func (f *fakeRepo) GetSpotCorrectionTarget(spotID string) (*models.SpotCorrectionTarget, error) {
	if f.target == nil || f.target.SpotID != spotID {
		return nil, nil
	}
	if f.saved != nil {
		copy := *f.target
		copy.Correction = f.saved
		if f.saved.SpotName != nil {
			copy.EffectiveSpotName = *f.saved.SpotName
		}
		if f.saved.PlaceID != nil {
			copy.EffectivePlaceID = *f.saved.PlaceID
		}
		if f.saved.Latitude != nil {
			copy.EffectiveLatitude = *f.saved.Latitude
		}
		if f.saved.Longitude != nil {
			copy.EffectiveLongitude = *f.saved.Longitude
		}
		if f.saved.YouTubeTimestampSeconds != nil {
			copy.EffectiveYouTubeTimestampSecs = f.saved.YouTubeTimestampSeconds
		}
		return &copy, nil
	}
	return f.target, nil
}

func (f *fakeRepo) UpsertSpotCorrection(c models.SpotCorrection) error {
	f.saved = &c
	return nil
}

type fakeLookup struct {
	coords *geocoder.Coordinates
	err    error
}

func (f fakeLookup) LookupPlaceIDCoordinates(context.Context, string) (*geocoder.Coordinates, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.coords, nil
}

type fakePlaceInputResolver struct {
	coords *geocoder.Coordinates
}

func (f fakePlaceInputResolver) LookupPlaceIDCoordinates(context.Context, string) (*geocoder.Coordinates, error) {
	return nil, errors.New("LookupPlaceIDCoordinates should not be called")
}

func (f fakePlaceInputResolver) ResolvePlaceInput(context.Context, string) (*geocoder.Coordinates, error) {
	return f.coords, nil
}

func TestSaveCorrectionResolvesPlaceAndPreservesPreviousOnFailure(t *testing.T) {
	repo := &fakeRepo{target: &models.SpotCorrectionTarget{SpotID: "1:1", EffectiveSpotName: "Old", EffectivePlaceID: "old"}}
	name := "New Name"
	placeID := "new_place"
	timestampURL := "https://youtu.be/abc?t=1m15s"
	_, err := NewService(repo, fakeLookup{coords: &geocoder.Coordinates{Latitude: 52.1, Longitude: 4.1}}).SaveCorrection(context.Background(), CorrectionInput{SpotID: "1:1", SpotName: &name, PlaceID: &placeID, TimestampedYouTubeURL: &timestampURL})
	if err != nil {
		t.Fatalf("SaveCorrection: %v", err)
	}
	if repo.saved == nil || repo.saved.PlaceID == nil || *repo.saved.PlaceID != placeID || repo.saved.Latitude == nil || *repo.saved.Latitude != 52.1 || repo.saved.YouTubeTimestampSeconds == nil || *repo.saved.YouTubeTimestampSeconds != 75 {
		t.Fatalf("unexpected saved correction: %+v", repo.saved)
	}

	previous := repo.saved
	_, err = NewService(repo, fakeLookup{err: errors.New("lookup failed")}).SaveCorrection(context.Background(), CorrectionInput{SpotID: "1:1", PlaceID: &placeID})
	if err == nil {
		t.Fatalf("expected lookup failure")
	}
	if repo.saved != previous {
		t.Fatalf("previous correction should remain unchanged on lookup failure")
	}
}

func TestSaveCorrectionCanHideSpot(t *testing.T) {
	repo := &fakeRepo{target: &models.SpotCorrectionTarget{SpotID: "1:1", EffectiveSpotName: "Old", EffectivePlaceID: "old"}}
	_, err := NewService(repo, fakeLookup{}).SaveCorrection(context.Background(), CorrectionInput{SpotID: "1:1", Hide: true})
	if err != nil {
		t.Fatalf("SaveCorrection: %v", err)
	}
	if repo.saved == nil || !repo.saved.Hidden {
		t.Fatalf("expected hidden correction, got %+v", repo.saved)
	}
}

func TestSaveCorrectionUsesResolvedPlaceIDFromAddressOrMapsURL(t *testing.T) {
	repo := &fakeRepo{target: &models.SpotCorrectionTarget{SpotID: "1:1", EffectiveSpotName: "Old", EffectivePlaceID: "old"}}
	placeInput := "https://maps.app.goo.gl/example"
	_, err := NewService(repo, fakePlaceInputResolver{coords: &geocoder.Coordinates{PlaceID: "ChIJresolved", Latitude: 52.1, Longitude: 4.1}}).SaveCorrection(context.Background(), CorrectionInput{SpotID: "1:1", PlaceID: &placeInput})
	if err != nil {
		t.Fatalf("SaveCorrection: %v", err)
	}
	if repo.saved == nil || repo.saved.PlaceID == nil || *repo.saved.PlaceID != "ChIJresolved" || repo.saved.Latitude == nil || *repo.saved.Latitude != 52.1 {
		t.Fatalf("unexpected saved correction: %+v", repo.saved)
	}
}

func TestSaveCorrectionUnknownSpot(t *testing.T) {
	_, err := NewService(&fakeRepo{}, fakeLookup{}).SaveCorrection(context.Background(), CorrectionInput{SpotID: "missing"})
	if !errors.Is(err, ErrSpotNotFound) {
		t.Fatalf("expected ErrSpotNotFound, got %v", err)
	}
}

func TestParseTimestampedYouTubeURLAcceptedForms(t *testing.T) {
	cases := map[string]int64{
		"https://youtube.com/watch?v=abc&t=90":       90,
		"https://www.youtube.com/watch?v=abc&t=1m2s": 62,
		"https://m.youtube.com/watch?v=abc&start=3":  3,
		"https://youtu.be/abc?t=1h2m3s":              3723,
	}
	for raw, want := range cases {
		got, err := ParseTimestampedYouTubeURL(raw)
		if err != nil || got != want {
			t.Fatalf("ParseTimestampedYouTubeURL(%q)=%d,%v want %d,nil", raw, got, err, want)
		}
	}
}

func TestParseTimestampedYouTubeURLRejectedForms(t *testing.T) {
	for _, raw := range []string{"90", "1:30", "https://example.com/watch?v=abc&t=90", "https://youtube.com/watch?v=abc", "https://youtu.be/abc"} {
		if _, err := ParseTimestampedYouTubeURL(raw); !errors.Is(err, ErrTimestampURLRequired) {
			t.Fatalf("expected timestamp URL error for %q, got %v", raw, err)
		}
	}
}
