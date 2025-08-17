package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

func helperValidateBody(body string) (string, error) {
	if len(body) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}

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
	return cleanWords, nil
}

func helperResponseError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	log.Printf("Error code %v: %s", code, msg)
	errorMsg := errorResponse{
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
