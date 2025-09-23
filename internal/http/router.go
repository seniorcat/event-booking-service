package http

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", PingHandler)
	mux.HandleFunc("/health", HealthHandler)

	return mux
}
