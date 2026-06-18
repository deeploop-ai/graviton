package shared

import (
	"context"
	"io"
	"time"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

type Queue interface {
	Enqueue(ctx context.Context, queue string, job any) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type BlobStore interface {
	Put(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	PresignedUploadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, map[string]string, error)
	PresignedDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
}
