package repository

import (
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func TestInitSchemaCreatesSpotCorrectionsTableIdempotently(t *testing.T) {
	repo := newTestRepo(t)
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("InitSchema second time: %v", err)
	}
	var name string
	if err := repo.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='spot_corrections'`).Scan(&name); err != nil {
		t.Fatalf("expected spot_corrections table: %v", err)
	}
}

func TestInitSchemaAddsSpotCorrectionsToExistingSchema(t *testing.T) {
	repo := newTestRepo(t)
	if _, err := repo.db.Exec(`DROP TABLE spot_corrections`); err != nil {
		t.Fatalf("drop correction table: %v", err)
	}
	sourceID, err := repo.UpsertArticleSource("https://example.com/existing")
	if err != nil {
		t.Fatalf("UpsertArticleSource: %v", err)
	}
	if err := repo.InitSchema(); err != nil {
		t.Fatalf("InitSchema migration: %v", err)
	}
	var count int
	if err := repo.db.QueryRow(`SELECT COUNT(*) FROM article_sources WHERE article_source_id=?`, sourceID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("existing source data not preserved, count=%d err=%v", count, err)
	}
	var name string
	if err := repo.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='spot_corrections'`).Scan(&name); err != nil {
		t.Fatalf("expected migrated spot_corrections table: %v", err)
	}
}

func TestSpotCorrectionSchemaRejectsInvalidRows(t *testing.T) {
	repo := newTestRepo(t)
	if _, err := repo.db.Exec(`INSERT INTO spot_corrections (spot_id, place_id) VALUES ('1:1', 'place_without_coords')`); err == nil {
		t.Fatalf("expected schema to reject place_id without coordinates")
	}
	if _, err := repo.db.Exec(`INSERT INTO spot_corrections (spot_id, youtube_timestamp_seconds) VALUES ('1:2', -1)`); err == nil {
		t.Fatalf("expected schema to reject negative timestamp")
	}
	if _, err := repo.db.Exec(`INSERT INTO spot_corrections (spot_id, hidden) VALUES ('1:3', 2)`); err == nil {
		t.Fatalf("expected schema to reject invalid hidden flag")
	}
}

func TestSpotCorrectionUpsertRetrieveAndPromptValues(t *testing.T) {
	repo := newTestRepo(t)
	seedExportFixture(t, repo)

	target, err := repo.GetSpotCorrectionTarget("1:1")
	if err != nil {
		t.Fatalf("GetSpotCorrectionTarget: %v", err)
	}
	if target == nil || target.SourceSpotName != "Stopera" || target.EffectiveSpotName != "Stopera" || target.EffectivePlaceID != "place_1" {
		t.Fatalf("unexpected source/effective target: %+v", target)
	}

	name := "Corrected"
	if err := repo.UpsertSpotCorrection(models.SpotCorrection{SpotID: "1:1", SpotName: &name}); err != nil {
		t.Fatalf("insert correction: %v", err)
	}
	correction, err := repo.GetSpotCorrection("1:1")
	if err != nil {
		t.Fatalf("GetSpotCorrection: %v", err)
	}
	if correction == nil || correction.SpotName == nil || *correction.SpotName != name {
		t.Fatalf("unexpected inserted correction: %+v", correction)
	}

	placeID := "place_2"
	lat := 1.2
	lng := 3.4
	if err := repo.UpsertSpotCorrection(models.SpotCorrection{SpotID: "1:1", SpotName: &name, PlaceID: &placeID, Latitude: &lat, Longitude: &lng}); err != nil {
		t.Fatalf("update correction: %v", err)
	}
	var rows int
	if err := repo.db.QueryRow(`SELECT COUNT(*) FROM spot_corrections WHERE spot_id='1:1'`).Scan(&rows); err != nil {
		t.Fatalf("count corrections: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected one correction row, got %d", rows)
	}

	hiddenCorrection, err := repo.GetSpotCorrection("1:1")
	if err != nil {
		t.Fatalf("GetSpotCorrection after update: %v", err)
	}
	if hiddenCorrection.Hidden {
		t.Fatalf("expected correction to be visible by default")
	}
	if err := repo.UpsertSpotCorrection(models.SpotCorrection{SpotID: "1:1", SpotName: &name, PlaceID: &placeID, Latitude: &lat, Longitude: &lng, Hidden: true}); err != nil {
		t.Fatalf("hide correction: %v", err)
	}
	hiddenCorrection, err = repo.GetSpotCorrection("1:1")
	if err != nil {
		t.Fatalf("GetSpotCorrection hidden: %v", err)
	}
	if !hiddenCorrection.Hidden {
		t.Fatalf("expected hidden correction: %+v", hiddenCorrection)
	}

	updated, err := repo.GetSpotCorrectionTarget("1:1")
	if err != nil {
		t.Fatalf("GetSpotCorrectionTarget updated: %v", err)
	}
	if updated.SourceSpotName != "Stopera" || updated.EffectiveSpotName != name || updated.EffectivePlaceID != placeID || updated.EffectiveLatitude != lat || updated.EffectiveLongitude != lng {
		t.Fatalf("source/effective values not correct: %+v", updated)
	}

	var sourcePlace string
	if err := repo.db.QueryRow(`SELECT place FROM spot_mentions WHERE spot_mention_id=1`).Scan(&sourcePlace); err != nil || sourcePlace != "Stopera" {
		t.Fatalf("source spot mention mutated: place=%q err=%v", sourcePlace, err)
	}
}
