package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func helperCleanBody(body string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	words := strings.Split(body, " ")
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if _, ok := badWords[lowerWord]; ok {
			words[i] = "****"
		}
	}
	cleanWords := strings.Join(words, " ")
	return cleanWords
}

func helperResponseError(w http.ResponseWriter, code int, msg string) {
	log.Printf("Error code %v: %s", code, msg)
	errorMsg := paramError{
		Error: msg,
	}
	helperResponseJSON(w, code, errorMsg)
}

func helperResponseJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}
