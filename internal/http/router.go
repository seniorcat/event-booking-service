package http

import (
	"net/http"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger"
	"laschool.ru/event-booking-service/internal/http/handlers"
	"laschool.ru/event-booking-service/internal/http/middleware"
	"laschool.ru/event-booking-service/internal/user"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	auth, err := middleware.NewAuthMiddleware()
	if err != nil {
		panic("failed to init auth middleware: " + err.Error())
	}

	mux.HandleFunc("/ping", handlers.PingHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	// Event CRUD
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListEvents(w, r)
		case http.MethodPost:
			auth(http.HandlerFunc(handlers.CreateEvent)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		// подпуть /events/{id}/bookings
		if strings.HasSuffix(r.URL.Path, "/bookings") && r.Method == http.MethodGet {
			handlers.ListBookingsByEvent(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.GetEvent(w, r)
		case http.MethodPut:
			auth(http.HandlerFunc(handlers.UpdateEvent)).ServeHTTP(w, r)
		case http.MethodDelete:
			auth(http.HandlerFunc(handlers.DeleteEvent)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Booking endpoints
	mux.HandleFunc("/bookings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			auth(http.HandlerFunc(handlers.CreateBooking)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/bookings/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetBooking(w, r)
		case http.MethodDelete:
			auth(http.HandlerFunc(handlers.CancelBooking)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// User endpoints
	mux.HandleFunc("/users/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			user.RegisterHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/users/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			user.LoginHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}
