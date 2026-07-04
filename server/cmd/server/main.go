package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arphost-com/Compose-Manager/server/internal/config"
	"github.com/arphost-com/Compose-Manager/server/internal/core"
	"github.com/arphost-com/Compose-Manager/server/internal/handlers"
	"github.com/arphost-com/Compose-Manager/server/internal/middleware"
	"github.com/arphost-com/Compose-Manager/server/internal/skills"
	"github.com/arphost-com/Compose-Manager/server/internal/skills/backup"
	"github.com/arphost-com/Compose-Manager/server/internal/skills/dbadmin"
	"github.com/arphost-com/Compose-Manager/server/internal/skills/debug"
	"github.com/arphost-com/Compose-Manager/server/internal/skills/frontend"
	"github.com/arphost-com/Compose-Manager/server/internal/skills/security"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Core engine
	engine := core.NewEngine(cfg.Root, cfg.HooksDir)

	// Skill registry
	registry := skills.NewRegistry()
	registry.Register(security.New())
	registry.Register(debug.New())
	registry.Register(backup.New())
	registry.Register(dbadmin.New())
	registry.Register(frontend.New())

	skillCfg := map[string]interface{}{
		"backup_dir": cfg.BackupDir,
	}

	ctx := context.Background()
	if err := registry.InitAll(ctx, engine, skillCfg); err != nil {
		log.Fatalf("skills init: %v", err)
	}

	// Handlers
	projectHandler := handlers.NewProjectHandler(engine)
	skillHandler := handlers.NewSkillHandler(registry)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(5 * time.Minute))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes (protected by API key)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.RequireAPIKey(cfg.APIKey))

		// Project endpoints
		r.Post("/projects", projectHandler.Create)
		r.Get("/projects", projectHandler.List)
		r.Get("/projects/{name}", projectHandler.Get)
		r.Get("/projects/{name}/images", projectHandler.Images)
		r.Get("/projects/{name}/status", projectHandler.Status)
		r.Post("/projects/{name}/pull", projectHandler.Pull)
		r.Post("/projects/{name}/up", projectHandler.Up)
		r.Post("/projects/{name}/down", projectHandler.Down)
		r.Post("/projects/{name}/update", projectHandler.Update)
		r.Post("/projects/{name}/restart", projectHandler.Restart)
		r.Put("/projects/{name}/inactive", projectHandler.SetInactive)

		// Bulk operations
		r.Post("/projects/bulk/{action}", projectHandler.BulkAction)

		// System
		r.Post("/prune", projectHandler.Prune)
		r.Post("/registries/login", projectHandler.RegistryLogin)

		// Skills
		r.Route("/skills", func(sr chi.Router) {
			sr.Get("/", skillHandler.List)
			sr.Get("/{skillName}", skillHandler.Get)
			registry.MountRoutes(sr)
		})
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Compose Manager API starting on %s (root: %s)", addr, cfg.Root)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	registry.ShutdownAll(shutCtx)
	srv.Shutdown(shutCtx)
	log.Println("server stopped")
}
