package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/models"
)

type correctionRepoStub struct {
	target *models.SpotCorrectionTarget
	saved  *models.SpotCorrection
}

func (r *correctionRepoStub) GetSpotCorrectionTarget(spotID string) (*models.SpotCorrectionTarget, error) {
	if r.target == nil || r.target.SpotID != spotID {
		return nil, nil
	}
	return r.target, nil
}

func (r *correctionRepoStub) UpsertSpotCorrection(c models.SpotCorrection) error {
	r.saved = &c
	return nil
}

func (r *correctionRepoStub) GetSpotCorrectionTargetForSource(_ models.SpotSource, spotID string) (*models.SpotCorrectionTarget, error) {
	return r.GetSpotCorrectionTarget(spotID)
}

func (r *correctionRepoStub) UpsertSpotCorrectionForSource(_ models.SpotSource, c models.SpotCorrection) error {
	return r.UpsertSpotCorrection(c)
}

type correctionLookupStub struct {
	err error
}

func (l correctionLookupStub) LookupPlaceIDCoordinates(context.Context, string) (*geocoder.Coordinates, error) {
	if l.err != nil {
		return nil, l.err
	}
	return &geocoder.Coordinates{Latitude: 52.2, Longitude: 4.2}, nil
}

func TestNormalizeCorrectionSpotIDArgument(t *testing.T) {
	cases := map[string]string{
		"2:20":                               "2:20",
		" 2:20 ":                             "2:20",
		"2%3A20":                             "2:20",
		"?spot=2%3A20":                       "2:20",
		"http://localhost:3000/?spot=2%3A20": "2:20",
		"http://localhost:3000/?foo=bar&spot=2%3A20": "2:20",
		"http://localhost:3000/#/map?spot=2%3A20":    "2:20",
	}
	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			got, err := normalizeCorrectionSpotIDArgument(raw)
			if err != nil {
				t.Fatalf("normalizeCorrectionSpotIDArgument: %v", err)
			}
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestNormalizeCorrectionSpotIDArgumentRejectsMissingSpotID(t *testing.T) {
	for _, raw := range []string{"", "?spot=", "?foo=bar", "http://localhost:3000/", "http://localhost:3000/?spot="} {
		if got, err := normalizeCorrectionSpotIDArgument(raw); err == nil {
			t.Fatalf("expected error for %q, got %q", raw, got)
		}
	}
}

func TestRunInteractiveSpotCorrectionSavesPromptedValues(t *testing.T) {
	ts := int64(15)
	repo := &correctionRepoStub{target: &models.SpotCorrectionTarget{
		SpotID:                        "1:1",
		EffectiveSpotName:             "Old Name",
		EffectivePlaceID:              "old_place",
		EffectiveYouTubeTimestampSecs: &ts,
	}}
	var out strings.Builder
	err := runInteractiveSpotCorrection(context.Background(), repo, correctionLookupStub{}, strings.NewReader("New Name\nnew_place\nhttps://youtu.be/abc?t=90\n"), &out, "1:1", models.SpotSourceGeminiDirect)
	if err != nil {
		t.Fatalf("runInteractiveSpotCorrection: %v", err)
	}
	if repo.saved == nil || repo.saved.SpotName == nil || *repo.saved.SpotName != "New Name" || repo.saved.PlaceID == nil || *repo.saved.PlaceID != "new_place" || repo.saved.YouTubeTimestampSeconds == nil || *repo.saved.YouTubeTimestampSeconds != 90 {
		t.Fatalf("unexpected saved correction: %+v", repo.saved)
	}
	if !strings.Contains(out.String(), "Current name: Old Name") || !strings.Contains(out.String(), "Saved correction") {
		t.Fatalf("expected current values and success output, got %q", out.String())
	}
}

func TestRunInteractiveSpotCorrectionBlankInputsPreserveExistingValues(t *testing.T) {
	repo := &correctionRepoStub{target: &models.SpotCorrectionTarget{SpotID: "1:1", EffectiveSpotName: "Old", EffectivePlaceID: "old"}}
	var out strings.Builder
	if err := runInteractiveSpotCorrection(context.Background(), repo, correctionLookupStub{}, strings.NewReader("\n\n\n"), &out, "1:1", models.SpotSourceGeminiDirect); err != nil {
		t.Fatalf("runInteractiveSpotCorrection: %v", err)
	}
	if repo.saved == nil || repo.saved.SpotName != nil || repo.saved.PlaceID != nil || repo.saved.YouTubeTimestampSeconds != nil {
		t.Fatalf("blank inputs should not create field overrides: %+v", repo.saved)
	}
}

func TestRunInteractiveSpotCorrectionActionableErrors(t *testing.T) {
	var out strings.Builder
	if err := runInteractiveSpotCorrection(context.Background(), &correctionRepoStub{}, correctionLookupStub{}, strings.NewReader("\n\n\n"), &out, "missing", models.SpotSourceGeminiDirect); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected unknown spot error, got %v", err)
	}

	repo := &correctionRepoStub{target: &models.SpotCorrectionTarget{SpotID: "1:1"}}
	if err := runInteractiveSpotCorrection(context.Background(), repo, correctionLookupStub{}, strings.NewReader("\n\n90\n"), &out, "1:1", models.SpotSourceGeminiDirect); err == nil || !strings.Contains(err.Error(), "timestamped youtube.com or youtu.be") {
		t.Fatalf("expected invalid timestamp error, got %v", err)
	}

	if err := runInteractiveSpotCorrection(context.Background(), repo, correctionLookupStub{err: errors.New("missing PRODUCTION_GOOGLE_MAPS_API_KEY")}, strings.NewReader("\nnew_place\n\n"), &out, "1:1", models.SpotSourceGeminiDirect); err == nil || !strings.Contains(err.Error(), "new_place") {
		t.Fatalf("expected place lookup error, got %v", err)
	}
}
