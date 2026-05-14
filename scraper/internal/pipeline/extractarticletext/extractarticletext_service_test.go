package extractarticletext

import (
	"context"
	"testing"
)

type fakeSQLite struct{}

type fakeFile struct{}

func (fakeSQLite) Run(_ context.Context, _ Request) (Response, error) {
	return Response{Stage: "sqlite"}, nil
}

func (fakeFile) Run(_ context.Context, _ Request) (Response, error) {
	return Response{Stage: "file"}, nil
}

func TestServiceRunNormalizesMode(t *testing.T) {
	svc := NewService(fakeSQLite{}, fakeFile{})

	resp, err := svc.Run(context.Background(), " SQLite ", Request{})
	if err != nil {
		t.Fatalf("Run sqlite: %v", err)
	}
	if resp.Stage != "sqlite" {
		t.Fatalf("expected sqlite stage, got %q", resp.Stage)
	}

	resp, err = svc.Run(context.Background(), " FILE ", Request{})
	if err != nil {
		t.Fatalf("Run file: %v", err)
	}
	if resp.Stage != "file" {
		t.Fatalf("expected file stage, got %q", resp.Stage)
	}
}
