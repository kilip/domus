package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrNotFound = errors.New("storage: object not found")
)

type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
	EnsureBucket(ctx context.Context) error
}
