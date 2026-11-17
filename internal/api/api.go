package api

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	chimMetrics "github.com/go-chi/metrics"
	"go.uber.org/zap"

	appconfig "github.com/IA-Generative/mozilla-extension-gate/internal/config"
	repositorywatcher "github.com/IA-Generative/mozilla-extension-gate/internal/repositoryWatcher"
)

const MetricPath = "/metrics"
const ReadyPath = "/readyz"
const HealthPath = "/healthz"

var SpecialPaths = []string{MetricPath, ReadyPath, HealthPath}

var safeTextRegex = regexp.MustCompile(`['";<>]`)

func sanitize(s string) string {
	return strings.TrimSpace(safeTextRegex.ReplaceAllString(s, ""))
}

func WriteSpec(cfg *appconfig.Config, log *zap.Logger) {
	router := chi.NewRouter()

	if err := os.MkdirAll("docs", 0o750); err != nil {
		log.Panic("failed to create docs directory", zap.Error(err))
	}

	f, err := os.Create("docs/openapi.json")
	if err != nil {
		log.Panic("failed to create spec file", zap.Error(err))
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err := enc.Encode(router); err != nil {
		log.Panic("failed to encode spec", zap.Error(err))
	}
}

func NewRouter(cfg *appconfig.Config, reconciler *repositorywatcher.RepositoryWatcher, logger *zap.Logger) http.Handler {
	// Initialize the r with the cfg.Server.BasePath as the base path for all routes
	base_r := chi.NewRouter()
	logger.Info("base path", zap.String("basePath", cfg.Server.BasePath))
	base_r.Route(cfg.Server.BasePath, func(r chi.Router) {

		r.Use(chimMetrics.Collector(chimMetrics.CollectorOpts{
			Host:  true,
			Proto: true,
			Skip: func(r *http.Request) bool {
				return slices.Contains(SpecialPaths, r.URL.Path)
			},
		}))

		r.Use(chimiddleware.RequestID)
		r.Use(chimiddleware.RealIP)
		r.Use(chimiddleware.Recoverer)
		r.Use(chimiddleware.Logger)

		r.Handle(MetricPath, chimMetrics.Handler())

		r.Get(HealthPath, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		r.Get(ReadyPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ready"}`))
		})

		r.Get("/releases", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			ctx := r.Context()
			list, err := reconciler.Repo.List(ctx, "")
			if err != nil {
				http.Error(w, "failed to list cached releases", http.StatusInternalServerError)
				return
			}
			body := map[string]any{
				"tags": list,
			}
			_ = json.NewEncoder(w).Encode(body)
		})

		r.Get("/updates.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			manifest := reconciler.GetUpdateManifest()
			if manifest == nil {
				http.Error(w, "failed to get update manifest", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(manifest)
		})

		if cfg.Server.ProxyArtifactEnabled {
			r.Get("/artifacts/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				logger.Info("id", zap.String("artifactId:", id))
				data, err := reconciler.GetArtifact(id)
				if err != nil {
					http.Error(w, "failed to download artifact", http.StatusInternalServerError)
					return
				}
				_, _ = w.Write(data)
			})
		}
	})

	transport := chimMetrics.Transport(chimMetrics.TransportOpts{Host: true})
	http.DefaultClient.Transport = transport(http.DefaultTransport)

	return base_r
}
