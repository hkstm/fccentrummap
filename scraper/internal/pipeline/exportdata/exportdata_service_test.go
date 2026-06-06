package exportdata

import (
	"context"
	"strings"
	"testing"

	"github.com/hkstm/fccentrummap/internal/models"
)

type captureSQLitePort struct{ req Request }

func (p *captureSQLitePort) Run(_ context.Context, req Request) (Response, error) {
	p.req = req
	return Response{Identity: "ok", Stage: "exportdata"}, nil
}

func TestServiceNormalizesSpotSourceForSQLite(t *testing.T) {
	port := &captureSQLitePort{}
	_, err := NewService(port, nil).Run(context.Background(), "sqlite", Request{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if port.req.SpotSource != models.SpotSourceGeminiDirect {
		t.Fatalf("default spot source = %q, want %q", port.req.SpotSource, models.SpotSourceGeminiDirect)
	}
	_, err = NewService(port, nil).Run(context.Background(), "sqlite", Request{SpotSource: models.SpotSourceGeminiDirect})
	if err != nil {
		t.Fatalf("Run gemini-direct: %v", err)
	}
	if port.req.SpotSource != models.SpotSourceGeminiDirect {
		t.Fatalf("gemini-direct spot source = %q", port.req.SpotSource)
	}
}

func TestServiceRejectsUnsupportedSpotSourceBeforeSQLiteAdapter(t *testing.T) {
	port := &captureSQLitePort{}
	_, err := NewService(port, nil).Run(context.Background(), "sqlite", Request{SpotSource: "bad"})
	if err == nil || !strings.Contains(err.Error(), "gemini-direct") {
		t.Fatalf("expected supported source error, got %v", err)
	}
	if port.req.SpotSource != "" {
		t.Fatalf("sqlite adapter should not have been called: %+v", port.req)
	}
}
