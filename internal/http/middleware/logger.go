package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method, path := r.Method, r.URL.Path
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		log.Printf("Method: %s, Path: %s, Code Response: %d, time: %v", method, path, rw.statusCode, time.Since(start))
	})
}
