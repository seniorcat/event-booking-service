package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"laschool.ru/event-booking-service/internal/booking"
	"laschool.ru/event-booking-service/internal/event"
	"laschool.ru/event-booking-service/pkg/container"
)

func parseID(path string) (int64, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// POST /bookings
func CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "container init failed"})
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	esvc := ctn.Get(event.DIEventService).(event.Service)

	var req struct {
		EventID int64 `json:"event_id"`
		UserID  int64 `json:"user_id"`
		Seats   int   `json:"seats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	e, err := esvc.Get(r.Context(), req.EventID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "event not found"})
		return
	}

	id, err := bsvc.Create(r.Context(), &booking.Booking{EventID: req.EventID, UserID: req.UserID, Seats: req.Seats}, e.Capacity)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// GET /bookings/{id}
func GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseID(r.URL.Path)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "container init failed"})
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	b, err := bsvc.Get(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// GET /events/{id}/bookings
func ListBookingsByEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// ожидаем /events/{id}/bookings
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[len(parts)-1] != "bookings" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	eventID, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event id"})
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "container init failed"})
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)

	limit, offset := 20, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			limit = p
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			offset = p
		}
	}

	list, err := bsvc.ListByEvent(r.Context(), eventID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list"})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// DELETE /bookings/{id}
func CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseID(r.URL.Path)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "container init failed"})
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	if err := bsvc.Cancel(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to cancel"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
