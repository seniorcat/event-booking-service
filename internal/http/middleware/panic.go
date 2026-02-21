package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)

				if _, err := w.Write([]byte("500 Internal Server Error")); err != nil {
					log.Printf("Failed to send 500 response: %v", err)
				}
			}
		}()
		next.ServeHTTP(w, r)

	})
}
