package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/seiobata/chirpy/internal/database"
)

const (
	rootPath = "."
	port     = "8080"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		db:       dbQueries,
		platform: platform,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(rootPath)))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerHitsMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	go func() {
		log.Printf("Starting server on port %s...\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// channel for shutdown
	quit := make(chan bool, 1)

	go func() {
		quitSig := make(chan os.Signal, 1)
		signal.Notify(quitSig, os.Interrupt, syscall.SIGTERM)
		<-quitSig
		quit <- true
	}()

	go func() {
		// slight delay so message prints after server starts
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Enter 'q' to shutdown server")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := scanner.Text()
			if input == "q" {
				quit <- true
				return
			}
			fmt.Println("Enter 'q' to shutdown server")
		}
	}()

	// wait for shutdown signal
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server failed to properly shutdown: %v", err)
	}
	log.Println("Server closed")
}
