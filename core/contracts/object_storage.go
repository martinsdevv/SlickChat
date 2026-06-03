package contracts

import (
	"context"
	"io"
	"time"
)

// ObjectStat is returned after upload confirmation.
type ObjectStat struct {
	Size        int64
	ContentType string
}

// ObjectStorage abstracts MinIO (S3-compatible) object lifecycle.
type ObjectStorage interface {
	EnsureBucket(ctx context.Context) error
	Delete(ctx context.Context, objectKey string) error
	Stat(ctx context.Context, objectKey string) (*ObjectStat, error)
	Get(ctx context.Context, objectKey string) (io.ReadCloser, *ObjectStat, error)
	Put(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) error
	PresignPut(ctx context.Context, objectKey, contentType string) (url string, expires time.Duration, err error)
	PresignGet(ctx context.Context, objectKey string) (url string, expires time.Duration, err error)
}
