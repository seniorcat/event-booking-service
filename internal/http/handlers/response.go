package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Status:  status,
		Message: v,
	}); err != nil {
		log.Printf("Failed to write error response: %м", err)
	}
}
