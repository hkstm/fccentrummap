package extractarticletext

import (
	"context"
	"fmt"
)

type FileAdapter struct{}

func NewFileAdapter() *FileAdapter { return &FileAdapter{} }

func (a *FileAdapter) Run(_ context.Context, _ Request) (Response, error) {
	return Response{}, fmt.Errorf("extractarticletext file adapter is not implemented; use --io sqlite")
}
