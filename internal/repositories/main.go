package repositories

import (
	"context"
	"io"

	"github.com/IA-Generative/mozilla-extension-gate/internal/config"
	file "github.com/IA-Generative/mozilla-extension-gate/internal/repositories/file"
	s3 "github.com/IA-Generative/mozilla-extension-gate/internal/repositories/s3"
	"go.uber.org/zap"
)

type Repository interface {
	Upload(ctx context.Context, key string, body []byte) error
	Download(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	GetFileLength(ctx context.Context, key string) (*int64, error)
	GetFile(ctx context.Context, key string) (*io.ReadCloser, error)
	Exists(ctx context.Context, key string) (bool, error)
}

func GetRepository(cfg config.Config, logger *zap.Logger) Repository {
	if cfg.FileStorage.Path != "" {
		return file.NewFileProvider(cfg)
	}

	r, err := s3.NewS3Provider(cfg.S3, logger)
	if err != nil {
		panic(err)
	}
	return r
}

func BuildKey(prefix, fileName string) string {
	return prefix + "/" + fileName
}
