package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

type storedItem struct {
	data        []byte
	contentType string
}

type MemoryStorage struct {
	mu      sync.RWMutex
	objects map[string]storedItem
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		objects: make(map[string]storedItem),
	}
}

func (m *MemoryStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = storedItem{
		data:        data,
		contentType: contentType,
	}
	return nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}

func (m *MemoryStorage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.objects[key]; !ok {
		return "", ErrNotFound
	}
	return fmt.Sprintf("http://mock-storage.local/%s", key), nil
}

func (m *MemoryStorage) EnsureBucket(ctx context.Context) error {
	return nil
}

func (m *MemoryStorage) Get(key string) ([]byte, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.objects[key]
	if !ok {
		return nil, "", ErrNotFound
	}
	return item.data, item.contentType, nil
}
