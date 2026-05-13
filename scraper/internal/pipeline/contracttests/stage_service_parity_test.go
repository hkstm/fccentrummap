package contracttests

import (
	"context"
	"testing"

	"github.com/hkstm/fccentrummap/internal/pipeline/acquireaudio"
	"github.com/hkstm/fccentrummap/internal/pipeline/collectarticleurls"
	"github.com/hkstm/fccentrummap/internal/pipeline/exportdata"
	"github.com/hkstm/fccentrummap/internal/pipeline/extractspots"
	"github.com/hkstm/fccentrummap/internal/pipeline/fetcharticles"
	"github.com/hkstm/fccentrummap/internal/pipeline/geocodespots"
	"github.com/hkstm/fccentrummap/internal/pipeline/transcribeaudio"
)

type caFake struct{ res collectarticleurls.Response }

func (f caFake) Run(context.Context, collectarticleurls.Request) (collectarticleurls.Response, error) {
	return f.res, nil
}

type faFake struct{ res fetcharticles.Response }

func (f faFake) Run(context.Context, fetcharticles.Request) (fetcharticles.Response, error) {
	return f.res, nil
}

type aaFake struct{ res acquireaudio.Response }

func (f aaFake) Run(context.Context, acquireaudio.Request) (acquireaudio.Response, error) {
	return f.res, nil
}

type taFake struct{ res transcribeaudio.Response }

func (f taFake) Run(context.Context, transcribeaudio.Request) (transcribeaudio.Response, error) {
	return f.res, nil
}

type esFake struct{ res extractspots.Response }

func (f esFake) Run(context.Context, extractspots.Request) (extractspots.Response, error) {
	return f.res, nil
}

type gsFake struct{ res geocodespots.Response }

func (f gsFake) Run(context.Context, geocodespots.Request) (geocodespots.Response, error) {
	return f.res, nil
}

type edFake struct{ res exportdata.Response }

func (f edFake) Run(context.Context, exportdata.Request) (exportdata.Response, error) {
	return f.res, nil
}

func TestStageServiceModeParityRouting(t *testing.T) {
	ctx := context.Background()
	if got, _ := collectarticleurls.NewService(caFake{collectarticleurls.Response{Stage: "s"}}, caFake{collectarticleurls.Response{Stage: "s"}}).Run(ctx, "sqlite", collectarticleurls.Request{}); got.Stage != "s" {
		t.Fatal("collect sqlite")
	}
	if got, _ := collectarticleurls.NewService(caFake{collectarticleurls.Response{Stage: "s"}}, caFake{collectarticleurls.Response{Stage: "s"}}).Run(ctx, "file", collectarticleurls.Request{}); got.Stage != "s" {
		t.Fatal("collect file")
	}

	if got, _ := fetcharticles.NewService(faFake{fetcharticles.Response{Stage: "s"}}, faFake{fetcharticles.Response{Stage: "s"}}).Run(ctx, "sqlite", fetcharticles.Request{}); got.Stage != "s" {
		t.Fatal("fetch sqlite")
	}
	if got, _ := fetcharticles.NewService(faFake{fetcharticles.Response{Stage: "s"}}, faFake{fetcharticles.Response{Stage: "s"}}).Run(ctx, "file", fetcharticles.Request{}); got.Stage != "s" {
		t.Fatal("fetch file")
	}

	if got, _ := acquireaudio.NewService(aaFake{acquireaudio.Response{Stage: "s"}}, aaFake{acquireaudio.Response{Stage: "s"}}).Run(ctx, "sqlite", acquireaudio.Request{}); got.Stage != "s" {
		t.Fatal("acquire sqlite")
	}
	if got, _ := acquireaudio.NewService(aaFake{acquireaudio.Response{Stage: "s"}}, aaFake{acquireaudio.Response{Stage: "s"}}).Run(ctx, "file", acquireaudio.Request{}); got.Stage != "s" {
		t.Fatal("acquire file")
	}

	if got, _ := transcribeaudio.NewService(taFake{transcribeaudio.Response{Stage: "s"}}, taFake{transcribeaudio.Response{Stage: "s"}}).Run(ctx, "sqlite", transcribeaudio.Request{}); got.Stage != "s" {
		t.Fatal("transcribe sqlite")
	}
	if got, _ := transcribeaudio.NewService(taFake{transcribeaudio.Response{Stage: "s"}}, taFake{transcribeaudio.Response{Stage: "s"}}).Run(ctx, "file", transcribeaudio.Request{}); got.Stage != "s" {
		t.Fatal("transcribe file")
	}

	if got, _ := extractspots.NewService(esFake{extractspots.Response{Stage: "s"}}, esFake{extractspots.Response{Stage: "s"}}).Run(ctx, "sqlite", extractspots.Request{}); got.Stage != "s" {
		t.Fatal("extract sqlite")
	}
	if got, _ := extractspots.NewService(esFake{extractspots.Response{Stage: "s"}}, esFake{extractspots.Response{Stage: "s"}}).Run(ctx, "file", extractspots.Request{}); got.Stage != "s" {
		t.Fatal("extract file")
	}

	if got, _ := geocodespots.NewService(gsFake{geocodespots.Response{Stage: "s"}}, gsFake{geocodespots.Response{Stage: "s"}}).Run(ctx, "sqlite", geocodespots.Request{}); got.Stage != "s" {
		t.Fatal("geocode sqlite")
	}
	if got, _ := geocodespots.NewService(gsFake{geocodespots.Response{Stage: "s"}}, gsFake{geocodespots.Response{Stage: "s"}}).Run(ctx, "file", geocodespots.Request{}); got.Stage != "s" {
		t.Fatal("geocode file")
	}

	if got, _ := exportdata.NewService(edFake{exportdata.Response{Stage: "s"}}, edFake{exportdata.Response{Stage: "s"}}).Run(ctx, "sqlite", exportdata.Request{}); got.Stage != "s" {
		t.Fatal("export sqlite")
	}
	if got, _ := exportdata.NewService(edFake{exportdata.Response{Stage: "s"}}, edFake{exportdata.Response{Stage: "s"}}).Run(ctx, "file", exportdata.Request{}); got.Stage != "s" {
		t.Fatal("export file")
	}
}
