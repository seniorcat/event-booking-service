package http

import (
	"net/http"
	"strings"

	"laschool.ru/event-booking-service/internal/http/handlers"
	"laschool.ru/event-booking-service/internal/user"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", handlers.PingHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Event CRUD
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListEvents(w, r)
		case http.MethodPost:
			handlers.CreateEvent(w, r)
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
			handlers.UpdateEvent(w, r)
		case http.MethodDelete:
			handlers.DeleteEvent(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Booking endpoints
	mux.HandleFunc("/bookings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateBooking(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/bookings/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetBooking(w, r)
		case http.MethodDelete:
			handlers.CancelBooking(w, r)
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

	return mux
}
