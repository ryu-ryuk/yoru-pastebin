package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	baseDir string
	baseURL string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(baseDir, baseURL string) (*LocalStorage, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

func (ls *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, info FileInfo) error {
	// Create the full file path
	filePath := filepath.Join(ls.baseDir, key)

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the data
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (ls *LocalStorage) GetDownloadURL(ctx context.Context, key string) (string, error) {
	// Check if file exists
	filePath := filepath.Join(ls.baseDir, key)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", key)
	}

	// For local storage, we'll use the paste ID to serve files securely
	// Extract paste ID from the key (format: uploads/pasteID/filename)
	parts := strings.Split(key, "/")
	if len(parts) >= 2 {
		pasteID := parts[1]
		return fmt.Sprintf("%s/%s/download", ls.baseURL, pasteID), nil
	}

	return "", fmt.Errorf("invalid key format: %s", key)
}

func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(ls.baseDir, key)
	return os.Remove(filePath)
}

func (ls *LocalStorage) IsAvailable(ctx context.Context) bool {
	// Check if we can write to the directory
	testFile := filepath.Join(ls.baseDir, ".test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}
