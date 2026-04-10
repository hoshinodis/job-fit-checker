package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/uhuko/job-fit-checker/backend/internal/api"
	"github.com/uhuko/job-fit-checker/backend/internal/config"
	"github.com/uhuko/job-fit-checker/backend/internal/db"
	"github.com/uhuko/job-fit-checker/backend/internal/llm"
	"github.com/uhuko/job-fit-checker/backend/internal/middleware"
	"github.com/uhuko/job-fit-checker/backend/internal/repository"
	"github.com/uhuko/job-fit-checker/backend/internal/service"
	"github.com/uhuko/job-fit-checker/backend/internal/worker"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	repo := repository.New(database)
	svc := service.New(repo)
	handler := api.NewHandler(svc)
	llmClient := llm.New(cfg.OllamaBaseURL, cfg.OllamaModel, cfg.RequestTimeoutSec)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.CORS)
	r.Use(middleware.MaxBodySize(1 * 1024 * 1024))

	rl := middleware.NewRateLimiter(cfg.RateLimitPerMinute)

	r.Get("/api/health", handler.Health)
	r.With(rl.Middleware).Post("/api/match", handler.PostMatch)
	r.Get("/api/match/{id}", handler.GetMatch)

	// Serve static files (Vue build output)
	fs := http.FileServer(http.Dir(cfg.StaticDir))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Try static file first; fall back to index.html for SPA routing
		path := cfg.StaticDir + req.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, req, cfg.StaticDir+"/index.html")
			return
		}
		fs.ServeHTTP(w, req)
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := worker.New(repo, llmClient, cfg)
	go w.Start(ctx)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("server starting on %s", addr)

	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
