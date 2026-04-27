package repository

import (
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

func TestResetSpotExtractionStorageWithBackupAndPresenterPersistence(t *testing.T) {
	repo := newTestRepo(t)
	articleRawID := insertTestArticleRaw(t, repo, "https://fccentrum.nl/story/de-spots-van-niels-oosthoek/")

	if err := repo.InsertArticleAudioSource(models.ArticleAudioSource{
		ArticleRawID: articleRawID,
		VideoID:      "video-1",
		YouTubeURL:   "https://youtube.com/watch?v=video-1",
		AudioFormat:  "mp3",
		MIMEType:     "audio/mpeg",
		AudioBlob:    []byte("audio"),
		ByteSize:     5,
	}); err != nil {
		t.Fatalf("insert audio source: %v", err)
	}
	audioSource, err := repo.GetArticleAudioSource(articleRawID)
	if err != nil || audioSource == nil {
		t.Fatalf("get audio source: row=%+v err=%v", audioSource, err)
	}

	transcriptionID, err := repo.UpsertArticleAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    audioSource.AudioSourceID,
		Provider:         "murmel",
		Language:         "nl",
		HTTPStatus:       200,
		ResponseJSON:     `{"segments":[{"text":"Stopera","start":1.2}]}`,
		ResponseByteSize: 42,
	})
	if err != nil {
		t.Fatalf("upsert transcription: %v", err)
	}

	presenter := "Niels Oosthoek"
	id1, err := repo.InsertSpotExtractionRecord(models.SpotExtractionRecordInput{
		ArticleRawID:       articleRawID,
		TranscriptionID:    transcriptionID,
		PresenterName:      &presenter,
		PromptText:         "prompt-1",
		RawResponseJSON:    `{"ok":true}`,
		ParsedResponseJSON: `{"presenter_name":"Niels Oosthoek","spots":[{"place":"Stopera","sentenceStartTimestamp":1.2}]}`,
	})
	if err != nil {
		t.Fatalf("insert extraction record: %v", err)
	}
	if id1 <= 0 {
		t.Fatalf("expected positive extraction id, got %d", id1)
	}

	backupTable, err := repo.ResetSpotExtractionStorageWithBackup("testbackup")
	if err != nil {
		t.Fatalf("reset extraction storage with backup: %v", err)
	}
	if backupTable == "" {
		t.Fatal("expected non-empty backup table name")
	}

	record, err := repo.GetLatestSpotExtractionRecord(articleRawID)
	if err != nil {
		t.Fatalf("get latest extraction after reset: %v", err)
	}
	if record != nil {
		t.Fatalf("expected no records after destructive reset, got %+v", record)
	}

	id2, err := repo.InsertSpotExtractionRecord(models.SpotExtractionRecordInput{
		ArticleRawID:       articleRawID,
		TranscriptionID:    transcriptionID,
		PromptText:         "prompt-2",
		RawResponseJSON:    `{"ok":true}`,
		ParsedResponseJSON: `{"spots":[{"place":"Oosterpark","sentenceStartTimestamp":3.4}]}`,
	})
	if err != nil {
		t.Fatalf("insert post-reset extraction record: %v", err)
	}
	if id2 <= 0 {
		t.Fatalf("expected positive extraction id after reset, got %d", id2)
	}

	record, err = repo.GetLatestSpotExtractionRecord(articleRawID)
	if err != nil {
		t.Fatalf("get latest extraction after reinsert: %v", err)
	}
	if record == nil {
		t.Fatal("expected extraction record after reinsert")
	}
	if record.PresenterName != nil {
		t.Fatalf("expected nullable presenter_name, got %+v", record.PresenterName)
	}
}

func TestResetSpotExtractionStorageWithBackupRejectsInvalidSuffix(t *testing.T) {
	repo := newTestRepo(t)

	if _, err := repo.ResetSpotExtractionStorageWithBackup("bad-suffix;DROP TABLE x"); err == nil {
		t.Fatal("expected invalid backup suffix error")
	}
}

func TestGetLatestArticleAudioTranscriptionByURL(t *testing.T) {
	repo := newTestRepo(t)
	articleRawID := insertTestArticleRaw(t, repo, "https://example.com/article-1")

	if err := repo.InsertArticleAudioSource(models.ArticleAudioSource{
		ArticleRawID: articleRawID,
		VideoID:      "video-2",
		YouTubeURL:   "https://youtube.com/watch?v=video-2",
		AudioFormat:  "mp3",
		MIMEType:     "audio/mpeg",
		AudioBlob:    []byte("audio"),
		ByteSize:     5,
	}); err != nil {
		t.Fatalf("insert audio source: %v", err)
	}
	audioSource, err := repo.GetArticleAudioSource(articleRawID)
	if err != nil || audioSource == nil {
		t.Fatalf("get audio source: row=%+v err=%v", audioSource, err)
	}

	_, err = repo.UpsertArticleAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    audioSource.AudioSourceID,
		Provider:         "murmel",
		Language:         "nl",
		HTTPStatus:       200,
		ResponseJSON:     `{"segments":[{"text":"eerste","start":1.1}]}`,
		ResponseByteSize: 10,
	})
	if err != nil {
		t.Fatalf("upsert first transcription: %v", err)
	}
	latestID, err := repo.UpsertArticleAudioTranscription(models.ArticleAudioTranscription{
		AudioSourceID:    audioSource.AudioSourceID,
		Provider:         "murmel-v2",
		Language:         "nl",
		HTTPStatus:       200,
		ResponseJSON:     `{"segments":[{"text":"tweede","start":2.2}]}`,
		ResponseByteSize: 10,
	})
	if err != nil {
		t.Fatalf("upsert second transcription: %v", err)
	}

	row, err := repo.GetLatestArticleAudioTranscriptionByURL("https://example.com/article-1")
	if err != nil {
		t.Fatalf("GetLatestArticleAudioTranscriptionByURL: %v", err)
	}
	if row == nil {
		t.Fatal("expected transcription row")
	}
	if row.TranscriptionID != latestID {
		t.Fatalf("expected latest transcription_id=%d got %d", latestID, row.TranscriptionID)
	}
}
