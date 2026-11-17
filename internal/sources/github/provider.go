package github

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IA-Generative/mozilla-extension-gate/internal/config"
	"go.uber.org/zap"
)

type GithubRelease struct {
	ID          int64                `json:"id"`
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Draft       bool                 `json:"draft"`
	Prerelease  bool                 `json:"prerelease"`
	PublishedAt time.Time            `json:"published_at"`
	HTMLURL     string               `json:"html_url"`
	AssetsURL   string               `json:"assets_url"`
	Assets      []GithubReleaseAsset `json:"assets"`
}
type GithubReleaseAsset struct {
	URL                string    `json:"url"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	ID                 int64     `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	State              string    `json:"state"`
	ContentType        string    `json:"content_type"`
	Size               int64     `json:"size"`
	Digest             string    `json:"digest"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		token: token,
	}
}

type GithubProvider struct {
	config *config.GithubConfig
	client *Client
	logger *zap.Logger
}

func NewGithubProvider(config *config.GithubConfig, logger *zap.Logger) (*GithubProvider, error) {
	if config == nil {
		logger.Fatal("GitHub config is required")
		return nil, fmt.Errorf("GitHub config is required")
	}
	if config.Repository == "" {
		return nil, nil
	}
	return &GithubProvider{
		config: config,
		client: NewClient(config.Token),
		logger: logger,
	}, nil
}
