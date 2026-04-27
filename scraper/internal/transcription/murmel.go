package transcription

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const defaultEndpoint = "https://api.murmel.eu/v1/transcribe"

type MurmelClient struct {
	apiKey   string
	endpoint string
	http     *http.Client
}

type TranscriptionResult struct {
	HTTPStatus int
	Body       []byte
	ErrMessage *string
}

func NewMurmelClient(apiKey string) *MurmelClient {
	return NewMurmelClientWithEndpoint(apiKey, defaultEndpoint)
}

func NewMurmelClientWithEndpoint(apiKey, endpoint string) *MurmelClient {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = defaultEndpoint
	}
	return &MurmelClient{
		apiKey:   apiKey,
		endpoint: endpoint,
		http:     &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *MurmelClient) Validate() error {
	if c.apiKey == "" {
		return fmt.Errorf("MURMEL_API_KEY is not set; set it in your environment before running transcribe-audio")
	}
	return nil
}

func (c *MurmelClient) Transcribe(ctx context.Context, filename string, audio []byte, language string) (*TranscriptionResult, error) {
	if language == "" {
		language = "nl"
	}

	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	filePart, err := writer.CreateFormFile("audio", filename)
	if err != nil {
		return nil, fmt.Errorf("creating multipart audio part: %w", err)
	}
	if _, err := filePart.Write(audio); err != nil {
		return nil, fmt.Errorf("writing multipart audio payload: %w", err)
	}
	if err := writer.WriteField("language", language); err != nil {
		return nil, fmt.Errorf("writing multipart language field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, &payload)
	if err != nil {
		return nil, fmt.Errorf("creating murmel request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		errText := fmt.Sprintf("murmel request failed: %v", err)
		return &TranscriptionResult{HTTPStatus: 0, Body: nil, ErrMessage: &errText}, nil
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		errText := fmt.Sprintf("reading murmel response failed: %v", readErr)
		return &TranscriptionResult{HTTPStatus: resp.StatusCode, Body: nil, ErrMessage: &errText}, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errText := fmt.Sprintf("murmel returned non-2xx status: %d", resp.StatusCode)
		return &TranscriptionResult{HTTPStatus: resp.StatusCode, Body: body, ErrMessage: &errText}, nil
	}

	return &TranscriptionResult{HTTPStatus: resp.StatusCode, Body: body}, nil
}
