package sources

import (
	"github.com/IA-Generative/mozilla-extension-gate/internal/config"
	"github.com/IA-Generative/mozilla-extension-gate/internal/sources/github"
	lib "github.com/IA-Generative/mozilla-extension-gate/internal/sources/lib"
	"go.uber.org/zap"
)

type SourceProvider interface {
	ListReleases() ([]lib.Release, error)
	ListReleaseAssets(release lib.Release) ([]lib.ReleaseAsset, error)
	FetchReleaseAsset(asset lib.ReleaseAsset) ([]byte, error)
}

func NewSourceProvider(cfg config.Config, logger *zap.Logger) SourceProvider {
	s, err := github.NewGithubProvider(&cfg.Github, logger)
	if err != nil {
		logger.Fatal("Failed to create GitHub provider", zap.Error(err))
	}
	if s == nil {
		return nil
	}

	return s
}
