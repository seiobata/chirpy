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
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(rootPath))))
	mux.HandleFunc("/healthz", handlerReadiness)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	log.Printf("Starting server on port %s...\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
