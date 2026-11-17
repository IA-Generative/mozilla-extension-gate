package sourcewatcher

import (
	"context"

	"github.com/IA-Generative/mozilla-extension-gate/internal/config"
	"github.com/IA-Generative/mozilla-extension-gate/internal/lib"
	"github.com/IA-Generative/mozilla-extension-gate/internal/repositories"
	repositorywatcher "github.com/IA-Generative/mozilla-extension-gate/internal/repositoryWatcher"
	"github.com/IA-Generative/mozilla-extension-gate/internal/sources"
	sourceProviderLib "github.com/IA-Generative/mozilla-extension-gate/internal/sources/lib"
	"go.uber.org/zap"
)

type SourceWatcher struct {
	provider    sources.SourceProvider
	Repo        repositories.Repository
	manifestSvc *repositorywatcher.RepositoryWatcher
	logger      *zap.Logger
	config      *config.Config
}

func NewSourceWatcher(provider sources.SourceProvider, repository repositories.Repository, logger *zap.Logger, config *config.Config) *SourceWatcher {
	return &SourceWatcher{
		Repo:     repository,
		provider: provider,
		logger:   logger,
		config:   config,
	}
}

func (s *SourceWatcher) Start() error {
	lib.RunPeriodically(s.config.Github.SyncInterval, s.Sync)
	// Implement the logic to start watching the source provider for changes
	return nil
}

func (s *SourceWatcher) Sync() error {
	s.logger.Info("Starting sync operation")
	releases, err := s.provider.ListReleases()
	if err != nil {
		s.logger.Error("Failed to list releases", zap.Error(err))
		return err
	}
	// hasChanged := false
	if len(releases) == 0 {
		s.logger.Info("No releases found")
		return nil
	}

	xpiAssetsList := []sourceProviderLib.ReleaseAsset{}
	for _, release := range releases {
		assets := release.Assets
		if assets == nil {
			assets, err := s.provider.ListReleaseAssets(release)
			if err != nil {
				s.logger.Error("Failed to list release assets", zap.Error(err))
				continue
			}
			release.Assets = &assets
		}
		if len(*release.Assets) == 0 {
			continue
		}
		xpiAssets := filterAssets(*release.Assets, func(asset sourceProviderLib.ReleaseAsset) bool {
			s.logger.Info("Processing asset", zap.Any("asset", asset))
			return asset.ContentType == "application/x-xpinstall"
		})
		if len(xpiAssets) == 0 {
			s.logger.Info("No XPI assets found for release", zap.Any("release", release.Name))
			continue
		}
		xpiAssetsList = append(xpiAssetsList, xpiAssets...)
	}
	s.logger.Info("Collected XPI assets", zap.Any("xpiAssets", xpiAssetsList))
	hasChanged := s.ProcessXPIAssets(xpiAssetsList)
	if hasChanged && s.manifestSvc != nil {
		s.manifestSvc.Sync()
		s.logger.Info("XPI assets have changed")
	}
	return nil
}

func filterAssets(assets []sourceProviderLib.ReleaseAsset, predicate func(sourceProviderLib.ReleaseAsset) bool) []sourceProviderLib.ReleaseAsset {
	filtered := []sourceProviderLib.ReleaseAsset{}
	for _, asset := range assets {
		if predicate(asset) {
			filtered = append(filtered, asset)
		}
	}
	return filtered
}

func (s *SourceWatcher) ProcessXPIAssets(xpiAssets []sourceProviderLib.ReleaseAsset) bool {
	anyChanges := false
	ctx := context.Background()
	for _, asset := range xpiAssets {
		key := repositories.BuildS3Key(s.config.S3.ArtifactsPrefix, asset.Name)
		exists, err := s.Repo.Exists(ctx, key)
		if err != nil {
			s.logger.Error("Failed to check if file exists", zap.Error(err))
			continue
		}
		if !exists {
			s.logger.Info("Asset not found in repository", zap.String("asset", asset.Name))
			s.SynchronizeAssetToRepository(asset, key)
			anyChanges = true
			continue
		}
		storedLength, err := s.Repo.GetFileLength(ctx, key)
		if err != nil || storedLength == nil {
			s.logger.Error("Failed to get file length", zap.Error(err))
			continue
		}
		s.logger.Info("Stored length for asset", zap.String("asset", asset.Name), zap.Int64("length", *storedLength))
		if *storedLength != asset.Size {
			s.SynchronizeAssetToRepository(asset, key)
			anyChanges = true
		}
	}
	return anyChanges
}

func (s *SourceWatcher) SynchronizeAssetToRepository(asset sourceProviderLib.ReleaseAsset, s3Key string) error {
	data, err := s.provider.FetchReleaseAsset(asset)
	if err != nil {
		s.logger.Error("Failed to fetch release asset", zap.String("asset", asset.Name), zap.Error(err))
		return err
	}
	err = s.Repo.Upload(context.Background(), s3Key, data)
	if err != nil {
		s.logger.Error("Failed to upload asset to repository", zap.String("asset", asset.Name), zap.Error(err))
		return err
	}
	s.logger.Info("Successfully uploaded asset to repository", zap.String("asset", asset.Name))
	return nil
}
