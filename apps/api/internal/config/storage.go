package config

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/paroki/domus/api/internal/platform/storage"
)

func GetStorage(cfg Config, logger *slog.Logger) storage.Storage {
	if cfg.S3Endpoint == "" {
		log.Fatalf("S3_ENDPOINT is required but not configured")
	}

	store, err := storage.NewS3Storage(storage.S3Config{
		Endpoint:       cfg.S3Endpoint,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		Region:         cfg.S3Region,
		ForcePathStyle: cfg.S3ForcePathStyle,
		Bucket:         cfg.S3Bucket,
	})
	if err != nil {
		log.Fatalf("failed to initialize S3 storage: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ensureBucketWithRetry(ctx, store, cfg.S3Bucket, logger); err != nil {
		log.Fatalf("failed to ensure S3 bucket %s: %v", cfg.S3Bucket, err)
	}

	return store
}

func ensureBucketWithRetry(ctx context.Context, store storage.Storage, bucket string, logger *slog.Logger) error {
	for i := 0; ; i++ {
		err := store.EnsureBucket(ctx)
		if err == nil {
			return nil
		}

		wait := time.Duration(1<<min(i, 4)) * time.Second
		if logger != nil {
			logger.Warn("waiting for S3 storage to become ready", "bucket", bucket, "retry_in", wait, "error", err)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for S3 bucket (%w): %w", ctx.Err(), err)
		case <-time.After(wait):
		}
	}
}
