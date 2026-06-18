package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ImageStorage interface {
	Save(ctx context.Context, key string, reader io.Reader) (string, error)
}

type LocalImageStorage struct {
	rootDir string
	baseURL string
}

func NewLocalImageStorage(rootDir string, baseURL string) *LocalImageStorage {
	return &LocalImageStorage{
		rootDir: rootDir,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *LocalImageStorage) Save(ctx context.Context, key string, reader io.Reader) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	cleanKey := filepath.Clean(strings.TrimLeft(key, `/\`))
	path := filepath.Join(s.rootDir, cleanKey)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", err
	}

	return s.baseURL + "/" + filepath.ToSlash(cleanKey), nil
}
