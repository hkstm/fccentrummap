package geocodespots

import (
	"context"
	"fmt"
)

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter { return &SQLiteAdapter{} }

func (a *SQLiteAdapter) Run(_ context.Context, _ Request) (Response, error) {
	return Response{}, fmt.Errorf("geocode-spots does not support --io sqlite yet; sqlite persistence is deferred. Use --io file --in <path>")
}
