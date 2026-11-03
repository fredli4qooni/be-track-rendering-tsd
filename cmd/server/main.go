package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"be-explore-maptalks/internal/database"
	"be-explore-maptalks/internal/entity"
	"be-explore-maptalks/internal/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/entities.db"
	}

	log.Println("==================================================")
	log.Println(" Tactical Situation Display (TSD) - Backend API ")
	log.Println(" Author: Fredli Fourqoni                          ")
	log.Println("==================================================")

	// 1. Initialize SQLite Database
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Database initialization failed: %v", err)
	}
	defer db.Close()

	// 2. Initialize Layered Architecture
	repo := entity.NewRepository(db)
	service := entity.NewService(repo)
	handler := entity.NewHandler(service)

	// 3. Pre-seed initial entities if empty
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if _, err := service.SeedInitialEntities(ctx); err != nil {
		log.Printf("[WARN] Auto-seed warning: %v", err)
	}
	cancel()

	// 4. Setup ServeMux & Routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 5. Apply Middlewares (Logger -> CORS -> Router)
	var rootHandler http.Handler = mux
	rootHandler = middleware.EnableCORS(rootHandler)
	rootHandler = middleware.RequestLogger(rootHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Start server in goroutine
	go func() {
		log.Printf("[Server] REST API listening on http://localhost:%s", port)
		log.Printf("[Server] Health Check available at http://localhost:%s/api/v1/health", port)
		log.Printf("[Server] Entities endpoint: http://localhost:%s/api/v1/entities", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server listening error: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[Server] Received signal %v. Initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[FATAL] Server forced to shutdown: %v", err)
	}

	log.Println("[Server] Server gracefully stopped.")
}
