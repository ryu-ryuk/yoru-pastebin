package storage

import (
	"context"
	"io"
)

// FileInfo represents metadata about an uploaded file
type FileInfo struct {
	Key          string
	OriginalName string
	Size         int64
	ContentType  string
}

// Storage interface for file storage operations
type Storage interface {
	// Upload stores a file and returns its key
	Upload(ctx context.Context, key string, reader io.Reader, info FileInfo) error

	// GetDownloadURL returns a URL to download the file
	GetDownloadURL(ctx context.Context, key string) (string, error)

	// Delete removes a file
	Delete(ctx context.Context, key string) error

	// IsAvailable checks if the storage backend is available
	IsAvailable(ctx context.Context) bool
}

// Config holds storage configuration
type Config struct {
	Type     string // "s3" or "local"
	LocalDir string // for local storage
	S3Config S3Config
}

type S3Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BaseURL         string // for constructing public URLs
}
