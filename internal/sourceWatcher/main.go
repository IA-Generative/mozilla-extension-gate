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

func (s *SourceWatcher) Start() {
	lib.RunPeriodically(s.config.Github.SyncInterval, s.Sync)
}

func (s *SourceWatcher) Sync() error {
	s.logger.Info("Starting sync source watcher")
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
			return asset.ContentType == "application/x-xpinstall"
		})
		if len(xpiAssets) == 0 {
			s.logger.Info("No XPI assets found for release", zap.Any("release", release.Name))
			continue
		}
		xpiAssetsList = append(xpiAssetsList, xpiAssets...)
	}
	xpiAssetsNameList := []string{}
	for _, asset := range xpiAssetsList {
		xpiAssetsNameList = append(xpiAssetsNameList, asset.Name)
	}
	s.logger.Info("Collected XPI assets", zap.Any("xpiAssetsNameList", xpiAssetsNameList))
	hasChanged := s.ProcessXPIAssets(xpiAssetsList)
	if hasChanged && s.manifestSvc != nil {
		err := s.manifestSvc.Sync()
		if err != nil {
			s.logger.Error("Failed to sync manifest service", zap.Error(err))
		}
		s.logger.Info("XPI assets have changed")
	}
	s.logger.Info("Finished sync source watcher")
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
		key := repositories.BuildKey(s.config.ArtifactsPrefix, asset.Name)
		exists, err := s.Repo.Exists(ctx, key)
		if err != nil {
			s.logger.Error("Failed to check if file exists", zap.Error(err))
			continue
		}
		if !exists {
			s.logger.Info("Asset not found in repository", zap.String("asset", asset.Name))
			err := s.SynchronizeAssetToRepository(asset, key)
			if err != nil {
				s.logger.Error("Failed to synchronize asset to repository", zap.String("asset", asset.Name), zap.Error(err))
			}
			anyChanges = true
			continue
		}
		storedLength, err := s.Repo.GetFileLength(ctx, key)
		if err != nil || storedLength == nil {
			s.logger.Error("Failed to get file length", zap.Error(err))
			continue
		}
		if *storedLength != asset.Size {
			s.logger.Info("Asset size mismatch, updating repository", zap.String("asset", asset.Name))
			err := s.SynchronizeAssetToRepository(asset, key)
			if err != nil {
				s.logger.Error("Failed to synchronize asset to repository", zap.String("asset", asset.Name), zap.Error(err))
			}
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
