package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IA-Generative/mozilla-extension-gate/internal/api"
	appconfig "github.com/IA-Generative/mozilla-extension-gate/internal/config"
	"github.com/IA-Generative/mozilla-extension-gate/internal/logger"
	"github.com/IA-Generative/mozilla-extension-gate/internal/repositories"
	repositorywatcher "github.com/IA-Generative/mozilla-extension-gate/internal/repositoryWatcher"
	sourcewatcher "github.com/IA-Generative/mozilla-extension-gate/internal/sourceWatcher"
	"github.com/IA-Generative/mozilla-extension-gate/internal/sources"
)

func main() {
	// xpi.ExtractManifest()
	cfg, err := appconfig.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Server.Env)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	repoProvider := repositories.GetRepository(*cfg, log)

	sources := sources.NewSourceProvider(*cfg, log)

	// If source provider is nil, don't start it and warn the it's disabled
	if sources == nil {
		log.Info("Source Provider not configured, no Github configuration provided")
	} else {
		srcSvc := sourcewatcher.NewSourceWatcher(sources, repoProvider, log, cfg)
		srcSvc.Start()
	}

	repoSvc := repositorywatcher.NewRepositoryWatcher(repoProvider, log, cfg)
	repoSvc.Start()

	// Application service
	router := api.NewRouter(cfg, repoSvc, log)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server starting", logger.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server stopped unexpectedly", logger.Error(err))
		}
	}()

	<-shutdown
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutCtx); err != nil {
		log.Error("graceful shutdown failed", logger.Error(err))
	}
	log.Info("server stopped")
}
