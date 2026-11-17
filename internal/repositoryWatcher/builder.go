package repositorywatcher

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/IA-Generative/mozilla-extension-gate/internal/xpi"
	"go.uber.org/zap"
)

type Application struct {
	Gecko struct {
		StrictMinVersion *string `json:"strict_min_version,omitempty"`
		StrictMaxVersion *string `json:"strict_max_version,omitempty"`
	} `json:"gecko"`
}

type UpdateElement struct {
	Version      string       `json:"version"`
	UpdateLink   string       `json:"update_link"`
	UpdateHash   *string      `json:"update_hash,omitempty"`
	Applications *Application `json:"applications,omitempty"`
}
type UpdateManifest struct {
	Addons map[string]struct {
		Updates []UpdateElement `json:"updates"`
	} `json:"addons"`
	Etag *string `json:"-"`
}

var currentManifest UpdateManifest

func (s *RepositoryWatcher) GetUpdateManifest() *UpdateManifest {
	return &currentManifest
}

func (s *RepositoryWatcher) BuildUpdateManifest() error {
	ctx := context.Background()
	manifestsKeys, err := s.Repo.List(ctx, s.config.ManifestsPrefix)
	if err != nil {
		s.logger.Error("Failed to list manifests", zap.Error(err))
		return err
	}
	var usedManifestKeys []string

	var newUpdateManifest UpdateManifest = UpdateManifest{
		Addons: make(map[string]struct {
			Updates []UpdateElement `json:"updates"`
		}),
	}
	// var manifests []xpi.Manifest
	for _, key := range manifestsKeys {
		// Fetch each manifest from the repository
		dataReadCloser, err := s.Repo.GetFile(ctx, key)
		if err != nil {
			s.logger.Error("Failed to fetch manifest", zap.String("key", key), zap.Error(err))
			continue
		}
		// transform the dataReadCloser to a byte slice
		data, err := io.ReadAll(*dataReadCloser)
		if err != nil {
			s.logger.Error("Failed to read manifest", zap.String("key", key), zap.Error(err))
			continue
		}

		var manifest xpi.Manifest
		// Unmarshal the manifest data into the manifests slice
		if err := json.Unmarshal(data, &manifest); err != nil {
			s.logger.Error("Failed to unmarshal manifest", zap.String("key", key), zap.Error(err))
			continue
		}
		// manifests = append(manifests, manifest)
		// Add the manifest to the new update manifest
		addonId := *manifest.BrowserSpecificSettings.Gecko.Id
		if addonId == "" {
			continue
		}
		addonId = fmt.Sprintf("{%s}", addonId)
		entry, ok := newUpdateManifest.Addons[addonId]
		if !ok {
			newUpdateManifest.Addons[addonId] = struct {
				Updates []UpdateElement `json:"updates"`
			}{}
			entry = newUpdateManifest.Addons[addonId]
		}
		// remove the .manifest.json from key

		artifactBaseName := strings.TrimSuffix(key, ".manifest.json")
		artifactBaseName = strings.TrimPrefix(artifactBaseName, s.config.ManifestsPrefix+"/")
		update := UpdateElement{
			Version:    manifest.Version,
			UpdateLink: s.buildUpdateLink(artifactBaseName),
		}

		entry.Updates = append(entry.Updates, update)
		usedManifestKeys = append(usedManifestKeys, key)
		newUpdateManifest.Addons[addonId] = entry
	}
	s.logger.Info("Fetched manifests", zap.Int("count", len(manifestsKeys)), zap.Int("used_count", len(usedManifestKeys)))
	etag := new(string)
	*etag = strings.Join(usedManifestKeys, ",")
	sha256Sum := sha256.Sum256([]byte(*etag))
	*etag = fmt.Sprintf("%x", sha256Sum[:])

	currentManifest = UpdateManifest{
		Addons: newUpdateManifest.Addons,
		Etag:   etag,
	}

	return nil
}

func (s *RepositoryWatcher) buildUpdateLink(artifactBaseName string) string {
	if s.config.Server.ProxyArtifactEnabled {
		return s.config.Server.PublicURL + "/" + s.config.ArtifactsPrefix + "/" + artifactBaseName
	}
	return s.config.Server.ArtifactBaseUrl + artifactBaseName
}
