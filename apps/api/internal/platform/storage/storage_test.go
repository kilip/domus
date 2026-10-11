package storage_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/paroki/domus/api/internal/platform/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	// EnsureBucket is a no-op that succeeds
	require.NoError(t, store.EnsureBucket(ctx))

	// Missing key PresignGet returns ErrNotFound
	_, err := store.PresignGet(ctx, "nonexistent", 5*time.Minute)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Missing key Get returns ErrNotFound
	_, _, err = store.Get("nonexistent")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Put an object
	key := "finance/1/tx-123/receipt.pdf"
	content := "dummy pdf content"
	err = store.Put(ctx, key, strings.NewReader(content), int64(len(content)), "application/pdf")
	require.NoError(t, err)

	// Get the object
	data, contentType, err := store.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "application/pdf", contentType)
	assert.Equal(t, []byte(content), data)

	// PresignGet returns deterministic mock URL
	url, err := store.PresignGet(ctx, key, 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "http://mock-storage.local/"+key, url)

	// Delete the object
	err = store.Delete(ctx, key)
	require.NoError(t, err)

	// PresignGet after delete returns ErrNotFound
	_, err = store.PresignGet(ctx, key, 5*time.Minute)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestS3Storage_MissingEndpoint(t *testing.T) {
	_, err := storage.NewS3Storage(storage.S3Config{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3_ENDPOINT is required")
}
