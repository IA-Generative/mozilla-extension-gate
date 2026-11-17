package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Github GithubConfig `yaml:"github"`
	S3     S3Config     `yaml:"s3"`
}

type GithubConfig struct {
	Repository   string        `yaml:"github_repository"`
	Token        string        `yaml:"github_token"`
	SyncInterval time.Duration `yaml:"github_sync_interval"`
}

type S3Config struct {
	AccessKeyID     string `yaml:"s3_access_key_id"`
	SecretAccessKey string `yaml:"s3_secret_access_key"`
	Bucket          string `yaml:"s3_bucket"`
	Region          string `yaml:"s3_region"`
	Endpoint        string `yaml:"s3_endpoint"`
	ArtifactsPrefix string `yaml:"s3_artifacts_prefix"`
	ManifestsPrefix string `yaml:"s3_manifests_prefix"`
}

type ServerConfig struct {
	Env                  string `yaml:"env"`
	Port                 string `yaml:"port"`
	PublicURL            string `yaml:"public_url"`
	Version              string `yaml:"version"`
	InstanceID           string `yaml:"instance_id"`
	ArtifactBaseUrl      string `yaml:"artifact_base_url"`
	ProxyArtifactEnabled bool   `yaml:"proxy_artifact_enabled"`
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
			InstanceID: "UNKNOWN",
		},
		Github: GithubConfig{
			SyncInterval: 5 * time.Minute,
			Token:        "",
		},
		S3: S3Config{
			Region: "us-east-1",
		},
	}
}

func applyEnvOverrides(config *Config) {
	config.Server.Env = getenv("ENV", config.Server.Env)
	config.Server.Port = getenv("PORT", config.Server.Port)
	config.Server.PublicURL = getenv("PUBLIC_URL", config.Server.PublicURL)
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
	config.Github.Token = getenv("GITHUB_TOKEN", config.Github.Token)
	config.Github.Repository = getenv("GITHUB_REPOSITORY", config.Github.Repository)
	config.Github.SyncInterval = getDuration("GITHUB_SYNC_INTERVAL", config.Github.SyncInterval)
	config.S3.AccessKeyID = getenv("S3_ACCESS_KEY_ID", config.S3.AccessKeyID)
	config.S3.SecretAccessKey = getenv("S3_SECRET_ACCESS_KEY", config.S3.SecretAccessKey)
	config.S3.Bucket = getenv("S3_BUCKET", config.S3.Bucket)
	config.S3.Region = getenv("S3_REGION", config.S3.Region)
	config.S3.Endpoint = getenv("S3_ENDPOINT", config.S3.Endpoint)
	config.S3.ArtifactsPrefix = getenv("S3_ARTIFACTS_PREFIX", config.S3.ArtifactsPrefix, "artifacts")
	config.S3.ManifestsPrefix = getenv("S3_MANIFESTS_PREFIX", config.S3.ManifestsPrefix, "manifests")
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
