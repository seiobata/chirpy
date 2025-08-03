package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	rootPath = "."
	port     = "8080"
)

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(rootPath)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerHitsMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsReset)

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
