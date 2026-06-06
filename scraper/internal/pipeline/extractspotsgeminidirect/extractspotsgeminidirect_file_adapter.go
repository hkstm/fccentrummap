package extractspotsgeminidirect

import (
	"context"
	"fmt"
)

type FileAdapter struct{}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(ctx context.Context, req Request) (Response, error) {
	_ = ctx
	_ = req
	return Response{}, fmt.Errorf("extract-spots-gemini-direct does not support --io file")
}
