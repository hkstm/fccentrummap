package collectarticleurls

import "context"

type SQLitePort interface {
	Run(ctx context.Context, req Request) (Response, error)
}

type FilePort interface {
	Run(ctx context.Context, req Request) (Response, error)
}
