package repositorywatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"

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

func (s *RepositoryWatcher) Start() {
	s.logger.Info("Starting watcher repository")
	lib.RunPeriodically(s.config.Github.SyncInterval, s.Sync)
}

func (s *RepositoryWatcher) Sync() error {
	s.logger.Info("Starting sync repository watcher")

	artifacts, err := s.ListArtifacts()
	if err != nil {
		s.logger.Error("Failed to list artifacts", zap.Error(err))
		return err
	}
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
	s.logger.Info("Finished sync repository watcher")

	return nil
}

func (s *RepositoryWatcher) ListManifests() ([]string, error) {
	ctx := context.Background()
	manifests, err := s.Repo.List(ctx, s.config.ManifestsPrefix)
	if err != nil {
		return nil, err
	}
	return manifests, nil
}

func (s *RepositoryWatcher) ListArtifacts() ([]string, error) {
	ctx := context.Background()
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
	artifactBase := path.Base(artifact)
	s.logger.Info("Creating artifact manifest", zap.String("artifact", artifactBase))
	ctx := context.Background()
	artifactPath := path.Join(s.config.ArtifactsPrefix, artifactBase)

	artifactData, err := s.Repo.Download(ctx, artifactPath)
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
	if err != nil {
		s.logger.Error("Failed to marshal manifest to JSON", zap.String("artifact", artifactBase), zap.Error(err))
		return err
	}
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

func (s *RepositoryWatcher) GetLatestArtifact(extensionID string) ([]byte, string, error) {
	ctx := context.Background()
	manifest := s.GetUpdateManifest()
	if manifest == nil {
		return nil, "", fmt.Errorf("update manifest not found")
	}
	latestArtifact, ok := manifest.Addons[extensionID]
	if !ok || len(latestArtifact.Updates) == 0 {
		return nil, "", fmt.Errorf("no artifact found for extension ID: %s", extensionID)
	}
	highestVersion, err := s.determineHigherVersion(latestArtifact)
	if err != nil {
		return nil, "", err
	}
	key := path.Join(s.config.ArtifactsPrefix, path.Base(highestVersion.UpdateLink))
	data, err := s.Repo.Download(ctx, key)
	if err != nil {
		s.logger.Error("Failed to download artifact", zap.String("artifact", key), zap.Error(err))
		return nil, "", err
	}
	return data, highestVersion.Version, nil
}

func (s *RepositoryWatcher) determineHigherVersion(updateList UpdateList) (UpdateElement, error) {
	if len(updateList.Updates) == 0 {
		return UpdateElement{}, fmt.Errorf("update list is empty")
	}
	highest := updateList.Updates[0]
	for _, update := range updateList.Updates[1:] {
		if compareVersions(update.Version, highest.Version) > 0 {
			highest = update
		}
	}
	return highest, nil
}

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] > parts2[i] {
			return 1
		} else if parts1[i] < parts2[i] {
			return -1
		}
	}
	if len(parts1) > len(parts2) {
		return 1
	} else if len(parts1) < len(parts2) {
		return -1
	}
	return 0
}
