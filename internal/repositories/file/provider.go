package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	appconfig "github.com/IA-Generative/mozilla-extension-gate/internal/config"
)

func NewFileProvider(cfg appconfig.Config) *FileProvider {
	// create the artifacts directory if it doesn't exist
	manifestsPath := path.Join(cfg.FileStorage.Path, cfg.ManifestsPrefix)
	artifactsPath := path.Join(cfg.FileStorage.Path, cfg.ArtifactsPrefix)

	if _, err := os.Stat(artifactsPath); os.IsNotExist(err) {
		os.MkdirAll(artifactsPath, 0755)
	}
	// create the manifests directory if it doesn't exist
	if _, err := os.Stat(manifestsPath); os.IsNotExist(err) {
		os.MkdirAll(manifestsPath, 0755)
	}
	return &FileProvider{
		path:          cfg.FileStorage.Path,
		manifestsPath: manifestsPath,
		artifactsPath: artifactsPath,
	}
}

type FileProvider struct {
	path          string
	manifestsPath string
	artifactsPath string
}

func (f *FileProvider) Upload(ctx context.Context, key string, body []byte) error {
	err := os.WriteFile(path.Join(f.path, key), body, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileProvider) Download(ctx context.Context, filePath string) ([]byte, error) {
	data, err := os.ReadFile(path.Join(f.path, filePath))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *FileProvider) List(ctx context.Context, prefix string) ([]string, error) {
	baseDir := path.Join(f.path, prefix)
	fmt.Println("Listing files with prefix:", baseDir)
	dirEntries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range dirEntries {
		entryName := entry.Name()
		fmt.Println(entryName)
		if !entry.IsDir() {
			files = append(files, entryName)
		}
	}
	return files, nil
}

func (f *FileProvider) GetFileLength(ctx context.Context, key string) (*int64, error) {
	info, err := os.Stat(path.Join(f.path, key))
	if err != nil {
		return nil, err
	}
	size := info.Size()
	return &size, nil
}
func (f *FileProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := os.Stat(path.Join(f.path, key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (f *FileProvider) GetFile(ctx context.Context, key string) (*io.ReadCloser, error) {
	file, err := os.Open(path.Join(f.path, key))
	if err != nil {
		return nil, err
	}

	// convert *os.File to *io.ReadCloser
	var readCloser io.ReadCloser = file

	return &readCloser, nil
}
func (f *FileProvider) Delete(ctx context.Context, key string) error {
	err := os.Remove(path.Join(f.path, key))
	if err != nil {
		return err
	}
	return nil
}

// Upload(ctx context.Context, key string, body []byte) error
// Download(ctx context.Context, key string) ([]byte, error)
// Delete(ctx context.Context, key string) error
// List(ctx context.Context, prefix string) ([]string, error)
// GetFileLength(ctx context.Context, key string) (*int64, error)
// GetFile(ctx context.Context, key string) (*io.ReadCloser, error)
// Exists(ctx context.Context, key string) (bool, error)
