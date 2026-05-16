package genai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	gogenai "google.golang.org/genai"
)

type Client struct {
	apiKey  string
	model   string
	baseURL string
	timeout time.Duration
	backend gogenai.Backend
}

type GenerateContentResult struct {
	StatusCode int
	Body       []byte
}

func NewClient(apiKey, model string) *Client {
	return NewClientWithEndpoint(apiKey, model, "")
}

func NewClientWithEndpoint(apiKey, model, endpointBase string) *Client {
	return &Client{
		apiKey:  strings.TrimSpace(apiKey),
		model:   strings.TrimSpace(model),
		baseURL: strings.TrimSpace(endpointBase),
		timeout: 180 * time.Second,
		backend: gogenai.BackendGeminiAPI,
	}
}

func (c *Client) Validate() error {
	if c.apiKey == "" {
		return fmt.Errorf("Gemini API key is not set; provide --gemini-api-key or set GEMINI_API_KEY/GOOGLE_API_KEY before running extract-spots-dry-run")
	}
	if c.model == "" {
		return fmt.Errorf("Gemma model is not configured; set --model or MODEL before running extract-spots-dry-run")
	}
	return nil
}

func (c *Client) GenerateContent(ctx context.Context, prompt string, config *gogenai.GenerateContentConfig) (*GenerateContentResult, error) {
	if config == nil {
		return nil, fmt.Errorf("generateContent config is required")
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}

	capture := &captureTransport{base: http.DefaultTransport}
	httpClient := &http.Client{Transport: capture, Timeout: c.timeout}

	cfg := &gogenai.ClientConfig{
		APIKey:     c.apiKey,
		Backend:    c.backend,
		HTTPClient: httpClient,
	}
	if c.baseURL != "" {
		cfg.HTTPOptions.BaseURL = c.baseURL
	}

	client, err := gogenai.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}

	contents := []*gogenai.Content{gogenai.NewContentFromText(prompt, gogenai.RoleUser)}
	response, err := client.Models.GenerateContent(ctx, c.model, contents, config)
	if err != nil {
		statusCode, body := capture.LastResponse()
		if statusCode > 0 {
			return nil, fmt.Errorf("generateContent returned non-2xx status %d: %s", statusCode, string(body))
		}
		return nil, fmt.Errorf("generateContent request failed: %w", err)
	}

	statusCode, body := capture.LastResponse()
	if statusCode == 0 {
		return nil, fmt.Errorf("generateContent response capture failed")
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("generateContent returned non-2xx status %d: %s", statusCode, string(body))
	}
	if len(body) == 0 {
		marshaled, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			return nil, fmt.Errorf("marshal generateContent response: %w", marshalErr)
		}
		body = marshaled
	}

	return &GenerateContentResult{StatusCode: statusCode, Body: body}, nil
}

type captureTransport struct {
	base http.RoundTripper

	mu             sync.Mutex
	lastStatusCode int
	lastBody       []byte
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		_ = resp.Body.Close()
		return nil, readErr
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	t.mu.Lock()
	t.lastStatusCode = resp.StatusCode
	t.lastBody = append([]byte(nil), body...)
	t.mu.Unlock()

	return resp, nil
}

func (t *captureTransport) LastResponse() (int, []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastStatusCode, append([]byte(nil), t.lastBody...)
}
