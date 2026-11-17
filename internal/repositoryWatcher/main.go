package repositorywatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/IA-Generative/mozilla-extension-gate/internal/config"
	"github.com/IA-Generative/mozilla-extension-gate/internal/lib"
	"github.com/IA-Generative/mozilla-extension-gate/internal/repositories"
	xpi "github.com/IA-Generative/mozilla-extension-gate/internal/xpi"
	"go.uber.org/zap"
)

type RepositoryWatcher struct {
	Repo   repositories.Repository
	logger *zap.Logger
	config *config.Config
}

func NewRepositoryWatcher(repo repositories.Repository, logger *zap.Logger, config *config.Config) *RepositoryWatcher {
	return &RepositoryWatcher{
		Repo:   repo,
		logger: logger,
		config: config,
	}
}

func (s *RepositoryWatcher) Start() error {
	s.logger.Info("Starting watcher repository")
	lib.RunPeriodically(s.config.Github.SyncInterval, s.Sync)
	return nil
}

func (s *RepositoryWatcher) Sync() error {
	artifacts, err := s.ListArtifacts()
	if err != nil {
		s.logger.Error("Failed to list artifacts", zap.Error(err))
		return err
	}
	s.logger.Info("Retrieved artifacts", zap.Any("artifacts", artifacts))
	manifests, err := s.ListManifests()
	if err != nil {
		s.logger.Error("Failed to list manifests", zap.Error(err))
		return err
	}
	s.logger.Info("Retrieved manifests", zap.Any("manifests", manifests))

	artifactsWithoutManifests := []string{}
	for _, artifact := range artifacts {
		hasManifest, err := s.CheckArtifactHasManifest(artifact, manifests)
		if err != nil {
			s.logger.Error("Failed to check if artifact has manifest", zap.String("artifact", artifact), zap.Error(err))
			continue
		}
		if !hasManifest {
			// s.logger.Warn("Artifact does not have a corresponding manifest", zap.String("artifact", artifact))
			artifactsWithoutManifests = append(artifactsWithoutManifests, artifact)
		}
	}
	if len(artifactsWithoutManifests) > 0 {
		s.logger.Info("Artifacts without corresponding manifests", zap.Any("artifacts", artifactsWithoutManifests))
	}
	for _, manifestKey := range artifactsWithoutManifests {
		err := s.CreateArtifactManifest(manifestKey)
		if err != nil {
			s.logger.Error("Failed to create artifact manifest", zap.String("artifact", manifestKey), zap.Error(err))
		}
	}
	err = s.BuildUpdateManifest()
	if err != nil {
		s.logger.Error("Failed to build update manifest", zap.Error(err))
		return err
	}
	s.logger.Info("Built update manifest")

	return nil
}

func (s *RepositoryWatcher) ListManifests() ([]string, error) {
	ctx := context.Background()
	manifests, err := s.Repo.List(ctx, s.config.ManifestsPrefix)
	if err != nil {
		s.logger.Error("Failed to list manifests", zap.Error(err))
		return nil, err
	}
	return manifests, nil
}

func (s *RepositoryWatcher) ListArtifacts() ([]string, error) {
	ctx := context.Background()
	s.logger.Info("Listing artifacts", zap.String("prefix", s.config.ArtifactsPrefix))
	artifacts, err := s.Repo.List(ctx, s.config.ArtifactsPrefix)
	if err != nil {
		return nil, err
	}
	filteredArtifacts := []string{}
	for _, key := range artifacts {
		if filepath.Ext(key) == ".xpi" {
			filteredArtifacts = append(filteredArtifacts, key)
		}
	}
	return filteredArtifacts, nil
}

func (s *RepositoryWatcher) CheckArtifactHasManifest(artifact string, manifests []string) (bool, error) {
	if slices.Contains(manifests, artifact+".manifest.json") {
		return true, nil
	}
	return false, nil
}

func (s *RepositoryWatcher) CreateArtifactManifest(artifact string) error {
	artifactBase := filepath.Base(artifact)
	s.logger.Info("Creating artifact manifest", zap.String("artifact", artifactBase))
	ctx := context.Background()
	artifactData, err := s.Repo.Download(ctx, artifact)
	if err != nil {
		s.logger.Error("Failed to download artifact", zap.String("artifact", artifact), zap.Error(err))
		return err
	}
	xpiManifest, err := xpi.ExtractManifestFromArchive(artifactData)
	if err != nil {
		s.logger.Error("Failed to extract manifest from artifact", zap.Error(err))
		return err
	}
	manifestKey := repositories.BuildKey(s.config.ManifestsPrefix, artifactBase+".manifest.json")
	jsonData, err := json.Marshal(xpiManifest)
	err = s.Repo.Upload(ctx, manifestKey, jsonData)
	if err != nil {
		s.logger.Error("Failed to create artifact manifest", zap.String("artifact", artifactBase), zap.Error(err))
		return err
	}
	s.logger.Info("Created artifact manifest", zap.String("artifact", artifactBase))
	return nil
}

func (s *RepositoryWatcher) GetArtifact(artifact string) ([]byte, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s/%s", s.config.ArtifactsPrefix, artifact)
	data, err := s.Repo.Download(ctx, key)
	if err != nil {
		s.logger.Error("Failed to download artifact", zap.String("artifact", key), zap.Error(err))
		return nil, err
	}
	return data, nil
}
