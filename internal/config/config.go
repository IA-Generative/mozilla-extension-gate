package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server          ServerConfig
	Github          GithubConfig
	S3              S3Config
	FileStorage     FileStorageConfig
	ArtifactsPrefix string
	ManifestsPrefix string
}

type GithubConfig struct {
	Repository            string
	Token                 string
	SyncInterval          time.Duration
	TLSInsecureSkipVerify bool
}

type S3Config struct {
	AccessKeyID          string
	SecretAccessKey      string
	Bucket               string
	Region               string
	Endpoint             string
	UsePathStyle         bool
	TLSInsecureSkipVerify bool
}

type FileStorageConfig struct {
	Path string
}

type ServerConfig struct {
	Env                  string
	Port                 string
	PublicURL            string
	BasePath             string
	Version              string
	InstanceID           string
	ArtifactBaseUrl      string
	ProxyArtifactEnabled bool
}

func Load() (*Config, error) {
	config := defaultConfig()
	applyEnvOverrides(config)

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Env:        "development",
			Port:       "8080",
			PublicURL:  "http://localhost:8080",
			BasePath:   "/",
			InstanceID: "UNKNOWN",
		},
		Github: GithubConfig{
			SyncInterval: 5 * time.Minute,
			Token:        "",
		},
		S3: S3Config{
			Region: "us-east-1",
			UsePathStyle: true,
		},
		ArtifactsPrefix: "artifacts",
		ManifestsPrefix: "manifests",
	}
}

func applyServerOverrides(config *Config) {
	config.Server.Env = getenv("ENV", config.Server.Env)
	config.Server.Port = getenv("PORT", config.Server.Port)
	config.Server.PublicURL = getenv("PUBLIC_URL", config.Server.PublicURL)
	config.Server.BasePath = getenv("BASE_PATH", config.Server.BasePath)
	config.Server.Version = getenv("APP_VERSION", config.Server.Version)
	config.Server.InstanceID = getenv("INSTANCE_ID", func() string {
		config.Server.ArtifactBaseUrl = getenv("ARTIFACT_BASE_URL", config.Server.ArtifactBaseUrl)
		config.Server.ProxyArtifactEnabled = getenv("PROXY_ARTIFACT_ENABLED", fmt.Sprintf("%v", config.Server.ProxyArtifactEnabled)) == "true"
		if v, err := os.Hostname(); err == nil {
			return v
		} else {
			return ""
		}
	}(), config.Server.InstanceID)
}

func applyS3Overrides(config *Config) {
	config.S3.AccessKeyID = getenv("S3_ACCESS_KEY_ID", config.S3.AccessKeyID)
	config.S3.SecretAccessKey = getenv("S3_SECRET_ACCESS_KEY", config.S3.SecretAccessKey)
	config.S3.Bucket = getenv("S3_BUCKET", config.S3.Bucket)
	config.S3.Region = getenv("S3_REGION", config.S3.Region)
	config.S3.Endpoint = getenv("S3_ENDPOINT", config.S3.Endpoint)
	config.S3.UsePathStyle = getenv("S3_USE_PATH_STYLE", fmt.Sprintf("%v", config.S3.UsePathStyle)) == "true"
	config.S3.TLSInsecureSkipVerify = getenv("TLS_INSECURE_SKIP_VERIFY", fmt.Sprintf("%v", config.S3.TLSInsecureSkipVerify)) == "true"
}

func applyGithubOverrides(config *Config) {
	config.Github.Token = getenv("GITHUB_TOKEN", config.Github.Token)
	config.Github.Repository = getenv("GITHUB_REPOSITORY", config.Github.Repository)
	config.Github.SyncInterval = getDuration("GITHUB_SYNC_INTERVAL", config.Github.SyncInterval)
	config.Github.TLSInsecureSkipVerify = getenv("TLS_INSECURE_SKIP_VERIFY", fmt.Sprintf("%v", config.Github.TLSInsecureSkipVerify)) == "true"
}

func applyArtifactsOverrides(config *Config) {
	config.ArtifactsPrefix = getenv("ARTIFACTS_PREFIX", config.ArtifactsPrefix)
	config.ManifestsPrefix = getenv("MANIFESTS_PREFIX", config.ManifestsPrefix)
}

func applyFileStorageOverrides(config *Config) {
	config.FileStorage.Path = getenv("FILE_STORAGE_PATH", config.FileStorage.Path)
}

func applyEnvOverrides(config *Config) {
	applyServerOverrides(config)
	applyS3Overrides(config)
	applyGithubOverrides(config)
	applyArtifactsOverrides(config)
	applyFileStorageOverrides(config)
}

func getDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

// Multi fallback getenv: checks primary env var, then first non-empty value from fallback list, then default fallback.
func getenv(key string, fallback ...string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	for _, fb := range fallback {
		if fb = strings.TrimSpace(fb); fb != "" {
			return fb
		}
	}
	return ""
}
