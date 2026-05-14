package audio

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

)

var acceptedAudioFormats = map[string]string{
	"wav":  "audio/wav",
	"m4a":  "audio/mp4",
	"mp3":  "audio/mpeg",
	"flac": "audio/flac",
	"ogg":  "audio/ogg",
	"webm": "audio/webm",
	"mp4":  "audio/mp4",
}

type DownloadedAudio struct {
	Path       string
	Format     string
	MIMEType   string
	YouTubeURL string
}

type AudioDownloader interface {
	Download(ctx context.Context, videoID string) (*DownloadedAudio, error)
}

type YTDLPDownloader struct{}

func (d *YTDLPDownloader) Download(ctx context.Context, videoID string) (*DownloadedAudio, error) {
	youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	tempDir, err := os.MkdirTemp("", "fccentrum-audio-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}

	success := false
	defer func() {
		if success {
			return
		}
		if removeErr := os.RemoveAll(tempDir); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Printf("WARN: failed to remove temp dir %s: %v", tempDir, removeErr)
		}
	}()

	outputTemplate := filepath.Join(tempDir, "%(id)s.%(ext)s")
	wavPath, wavErr := runYTDLP(ctx, outputTemplate, youtubeURL, true)
	if wavErr == nil {
		success = true
		return &DownloadedAudio{
			Path:       wavPath,
			Format:     "wav",
			MIMEType:   acceptedAudioFormats["wav"],
			YouTubeURL: youtubeURL,
		}, nil
	}

	fallbackPath, fallbackErr := runYTDLP(ctx, outputTemplate, youtubeURL, false)
	if fallbackErr != nil {
		return nil, fmt.Errorf("wav download failed: %w; fallback failed: %w", wavErr, fallbackErr)
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fallbackPath)), ".")
	mimeType, ok := acceptedAudioFormats[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported downloaded audio format: %s", ext)
	}

	success = true
	return &DownloadedAudio{
		Path:       fallbackPath,
		Format:     ext,
		MIMEType:   mimeType,
		YouTubeURL: youtubeURL,
	}, nil
}

func runYTDLP(ctx context.Context, outputTemplate, youtubeURL string, preferWAV bool) (string, error) {
	args := []string{"--no-playlist", "-f", "bestaudio", "-o", outputTemplate, "--print", "after_move:filepath"}
	if preferWAV {
		args = append(args, "-x", "--audio-format", "wav")
	}
	args = append(args, youtubeURL)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running yt-dlp: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if _, statErr := os.Stat(line); statErr == nil {
			return line, nil
		}
	}

	return "", fmt.Errorf("yt-dlp did not report an output file")
}

func cleanupTempAudio(path string) {
	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		log.Printf("WARN: failed to remove temp audio file %s: %v", path, removeErr)
	}
	dir := filepath.Dir(path)
	if dirErr := os.Remove(dir); dirErr != nil && !os.IsNotExist(dirErr) {
		log.Printf("WARN: failed to remove temp audio dir %s: %v", dir, dirErr)
	}
}

