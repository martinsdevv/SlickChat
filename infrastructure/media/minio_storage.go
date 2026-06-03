package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/infrastructure/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	uploadTTL time.Duration
	readTTL   time.Duration
}

func (MinIOStorage) Configured() bool { return true }

func NewMinIOStorage(cfg config.MediaConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, err
	}

	uploadTTL := time.Duration(cfg.PresignUploadTTL) * time.Second
	if uploadTTL <= 0 {
		uploadTTL = 15 * time.Minute
	}
	readTTL := time.Duration(cfg.PresignReadTTL) * time.Second
	if readTTL <= 0 {
		readTTL = time.Hour
	}

	return &MinIOStorage{
		client:    client,
		bucket:    cfg.Bucket,
		uploadTTL: uploadTTL,
		readTTL:   readTTL,
	}, nil
}

func (s *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func (s *MinIOStorage) Delete(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return nil
	}
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) Stat(ctx context.Context, objectKey string) (*contracts.ObjectStat, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return objectStatFromInfo(info), nil
}

func (s *MinIOStorage) Get(ctx context.Context, objectKey string) (io.ReadCloser, *contracts.ObjectStat, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, nil, err
	}
	return obj, objectStatFromInfo(info), nil
}

func objectStatFromInfo(info minio.ObjectInfo) *contracts.ObjectStat {
	return &contracts.ObjectStat{
		Size:        info.Size,
		ContentType: info.ContentType,
	}
}

func (s *MinIOStorage) Put(
	ctx context.Context,
	objectKey, contentType string,
	body io.Reader,
	size int64,
) error {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinIOStorage) PresignPut(ctx context.Context, objectKey, contentType string) (string, time.Duration, error) {
	// PresignedPutObject only signs "host"; PUT with Content-Type must use PresignHeader.
	headers := make(http.Header)
	if contentType != "" {
		headers.Set("Content-Type", contentType)
	}
	u, err := s.client.PresignHeader(
		ctx,
		http.MethodPut,
		s.bucket,
		objectKey,
		s.uploadTTL,
		url.Values{},
		headers,
	)
	if err != nil {
		return "", 0, err
	}
	return u.String(), s.uploadTTL, nil
}

func (s *MinIOStorage) PresignGet(ctx context.Context, objectKey string) (string, time.Duration, error) {
	reqParams := make(url.Values)
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, s.readTTL, reqParams)
	if err != nil {
		return "", 0, err
	}
	return u.String(), s.readTTL, nil
}

// NewObjectStorageFromConfig returns MinIO storage or NoopStorage on connection failure.
func NewObjectStorageFromConfig(cfg config.MediaConfig) contracts.ObjectStorage {
	store, err := NewMinIOStorage(cfg)
	if err != nil {
		return NoopStorage{}
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		return NoopStorage{}
	}
	return store
}

func MustMinIOStorage(cfg config.MediaConfig) (*MinIOStorage, error) {
	store, err := NewMinIOStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("minio bucket: %w", err)
	}
	return store, nil
}
