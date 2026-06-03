package media

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/martinsdevv/slickchat/core/contracts"
)

// NoopStorage is used when MinIO is unavailable.
type NoopStorage struct{}

func (NoopStorage) Configured() bool { return false }

func (NoopStorage) EnsureBucket(context.Context) error { return nil }

func (NoopStorage) Delete(_ context.Context, _ string) error { return nil }

func (NoopStorage) Stat(context.Context, string) (*contracts.ObjectStat, error) {
	return nil, errors.New("object storage not configured")
}

func (NoopStorage) Get(context.Context, string) (io.ReadCloser, *contracts.ObjectStat, error) {
	return nil, nil, errors.New("object storage not configured")
}

func (NoopStorage) Put(context.Context, string, string, io.Reader, int64) error {
	return errors.New("object storage not configured")
}

func (NoopStorage) PresignPut(context.Context, string, string) (string, time.Duration, error) {
	return "", 0, errors.New("object storage not configured")
}

func (NoopStorage) PresignGet(context.Context, string) (string, time.Duration, error) {
	return "", 0, errors.New("object storage not configured")
}
