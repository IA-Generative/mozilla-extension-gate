package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
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

var ignoredPaths []string

func genrateIgnoredPaths(skipPaths []string, prefix string) {
	ignoredPaths = []string{}
	for _, path := range skipPaths {
		ignoredPaths = append(ignoredPaths, prefix+path)
	}
}

func LoggingMiddlewareWithSkip(skipPaths []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(skipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			chimiddleware.Logger(next).ServeHTTP(w, r)
		})
	}
}

func NewRouter(cfg *appconfig.Config, reconciler *repositorywatcher.RepositoryWatcher, logger *zap.Logger) http.Handler {
	// Initialize the r with the cfg.Server.BasePath as the base path for all routes
	genrateIgnoredPaths(SpecialPaths, cfg.Server.BasePath)
	base_r := chi.NewRouter()
	base_r.Use(LoggingMiddlewareWithSkip(ignoredPaths))

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

		extensionAliasMap := make(map[string]string)
		if cfg.Server.ExtensionAlias != "" {
			aliases := strings.Split(cfg.Server.ExtensionAlias, ",")
			for _, alias := range aliases {
				parts := strings.Split(alias, ":")
				if len(parts) == 2 {
					extensionAliasMap[parts[0]] = parts[1]
				}
			}
		}
		r.Get("/{alias}/latest", func(w http.ResponseWriter, r *http.Request) {
			alias := chi.URLParam(r, "alias")
			geckoID, ok := extensionAliasMap[alias]
			if !ok {
				http.Error(w, "alias not found", http.StatusNotFound)
				return
			}
			data, err := reconciler.GetLatestArtifact(geckoID)
			if err != nil {
				http.Error(w, "failed to download artifact", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(chi.URLParam(r, "alias")+".xpi")))
			w.Header().Set("Content-Type", "application/x-xpinstall")
			_, _ = w.Write(data)
		})
	})

	transport := chimMetrics.Transport(chimMetrics.TransportOpts{Host: true})
	http.DefaultClient.Transport = transport(http.DefaultTransport)

	return base_r
}
