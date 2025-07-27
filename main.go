package main

import (
	"log"
	"net/http"
)

const (
	rootPath = "."
	port     = "8080"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(rootPath)))

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	log.Printf("Starting server on port %s...\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
