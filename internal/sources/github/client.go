package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	lib "github.com/IA-Generative/mozilla-extension-gate/internal/sources/lib"
	"go.uber.org/zap"
)

func (p *GithubProvider) ListReleases() ([]lib.Release, error) {
	endpoint := fmt.Sprintf(
		"https://api.github.com/repos/%s/releases?per_page=100&page=1",
		p.config.Repository,
	)
	p.logger.Info("Fetching releases from GitHub endpoint", zap.String("endpoint", endpoint))

	req, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		p.logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if p.client.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.client.token)
	}

	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		p.logger.Error("Failed to execute request", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err := resp.Body.Close()
		if err != nil {
			p.logger.Error("Failed to close response body", zap.Error(err))
		}
		return nil, fmt.Errorf(
			"GitHub API returned status %d",
			resp.StatusCode,
		)
	}

	var releases []GithubRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		p.logger.Error("Failed to decode GitHub releases", zap.Error(err))
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		p.logger.Error("Failed to close response body", zap.Error(err))
	}

	var filteredReleases []GithubRelease
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		filteredReleases = append(filteredReleases, release)
	}
	return convertGithubReleasesToLibReleases(filteredReleases), nil
}

func (p *GithubProvider) ListReleaseAssets(release lib.Release) ([]lib.ReleaseAsset, error) {
	endpoint := release.AssetsURL
	p.logger.Info("Fetching releases assets from GitHub endpoint", zap.String("endpoint", endpoint))

	req, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		p.logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if p.client.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.client.token)
	}

	p.logger.Info("Fetching release assets from GitHub endpoint", zap.String("endpoint", endpoint))
	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		p.logger.Error("Failed to execute request", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err := resp.Body.Close()
		if err != nil {
			p.logger.Error("Failed to close response body", zap.Error(err))
		}
		return nil, fmt.Errorf(
			"GitHub API returned status %d",
			resp.StatusCode,
		)
	}

	var assets []GithubReleaseAsset
	err = json.NewDecoder(resp.Body).Decode(&assets)

	if err != nil {
		p.logger.Error("Failed to decode GitHub release assets", zap.Error(err))
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		p.logger.Error("Failed to close response body", zap.Error(err))
	}

	return convertGithubReleaseAssetsToLibReleaseAssets(assets), nil
}

// Download the content of a release asset from GitHub
func (p *GithubProvider) FetchReleaseAsset(asset lib.ReleaseAsset) ([]byte, error) {
	// Use the API URL with Accept: application/octet-stream so GitHub returns a
	// redirect to a signed CDN URL. This avoids Go stripping the Authorization
	// header when following the cross-host redirect that browser_download_url triggers.
	endpoint := asset.URL
	p.logger.Info("Fetching release asset from GitHub endpoint", zap.String("endpoint", endpoint), zap.String("asset", asset.Name))

	req, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		p.logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if p.client.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.client.token)
	}

	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		p.logger.Error("Failed to execute request", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err := resp.Body.Close()
		if err != nil {
			p.logger.Error("Failed to close response body", zap.Error(err))
		}
		return nil, fmt.Errorf(
			"GitHub API returned status %d",
			resp.StatusCode,
		)
	}

	return io.ReadAll(resp.Body)
}

func convertGithubReleasesToLibReleases(releases []GithubRelease) []lib.Release {
	libReleases := make([]lib.Release, len(releases))
	for i, r := range releases {
		assets := convertGithubReleaseAssetsToLibReleaseAssets(r.Assets)
		libReleases[i] = lib.Release{
			ID:     strconv.FormatInt(r.ID, 10),
			Name:   r.Name,
			Assets: &assets,
		}
	}
	return libReleases
}

func convertGithubReleaseAssetsToLibReleaseAssets(assets []GithubReleaseAsset) []lib.ReleaseAsset {
	libAssets := make([]lib.ReleaseAsset, len(assets))
	for i, a := range assets {
		libAssets[i] = lib.ReleaseAsset{
			ID:          fmt.Sprintf("%d", a.ID),
			Name:        a.Name,
			URL:         a.URL,
			ContentType: a.ContentType,
			Size:        a.Size,
		}
	}
	return libAssets
}
