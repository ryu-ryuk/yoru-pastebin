package storage

import (
	"context"
	"fmt"
	"io"
	"log"
)

// HybridStorage provides fallback between S3 and local storage
type HybridStorage struct {
	primary    Storage
	fallback   Storage
	usePrimary bool
}

// NewHybridStorage creates a hybrid storage that tries S3 first, falls back to local
func NewHybridStorage(s3Config S3Config, localDir, baseURL string) (*HybridStorage, error) {
	// Initialize S3 storage
	s3Storage, err := NewS3Storage(s3Config)
	if err != nil {
		log.Printf("Failed to initialize S3 storage: %v", err)
	}

	// Initialize local storage
	localStorage, err := NewLocalStorage(localDir, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local storage: %w", err)
	}

	hs := &HybridStorage{
		fallback: localStorage,
	}

	// Test S3 availability
	if s3Storage != nil {
		ctx := context.Background()
		if s3Storage.IsAvailable(ctx) {
			hs.primary = s3Storage
			hs.usePrimary = true
			log.Println("Using S3 storage as primary")
		} else {
			log.Println("S3 not available, using local storage only")
		}
	} else {
		log.Println("S3 not configured, using local storage only")
	}

	return hs, nil
}

func (hs *HybridStorage) Upload(ctx context.Context, key string, reader io.Reader, info FileInfo) error {
	if hs.usePrimary && hs.primary != nil {
		err := hs.primary.Upload(ctx, key, reader, info)
		if err == nil {
			return nil
		}
		log.Printf("Primary storage failed, falling back to local: %v", err)
		hs.usePrimary = false
	}

	return hs.fallback.Upload(ctx, key, reader, info)
}

func (hs *HybridStorage) GetDownloadURL(ctx context.Context, key string) (string, error) {
	if hs.usePrimary && hs.primary != nil {
		url, err := hs.primary.GetDownloadURL(ctx, key)
		if err == nil {
			return url, nil
		}
		log.Printf("Primary storage failed for download, trying fallback: %v", err)
	}

	return hs.fallback.GetDownloadURL(ctx, key)
}

func (hs *HybridStorage) Delete(ctx context.Context, key string) error {
	var lastErr error

	// Try to delete from both storages
	if hs.primary != nil {
		if err := hs.primary.Delete(ctx, key); err != nil {
			lastErr = err
		}
	}

	if err := hs.fallback.Delete(ctx, key); err != nil {
		lastErr = err
	}

	return lastErr
}

func (hs *HybridStorage) IsAvailable(ctx context.Context) bool {
	if hs.usePrimary && hs.primary != nil {
		return hs.primary.IsAvailable(ctx)
	}
	return hs.fallback.IsAvailable(ctx)
}

// ForceLocal switches to local storage only
func (hs *HybridStorage) ForceLocal() {
	hs.usePrimary = false
	log.Println("Switched to local storage only")
}

// GetStorageInfo returns information about current storage mode
func (hs *HybridStorage) GetStorageInfo() string {
	if hs.usePrimary {
		return "S3 (with local fallback)"
	}
	return "Local storage"
}
