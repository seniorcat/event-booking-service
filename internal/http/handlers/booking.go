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
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "container init failed")
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
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	e, err := esvc.Get(r.Context(), req.EventID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "event not found")
		return
	}

	id, err := bsvc.Create(r.Context(), &booking.Booking{EventID: req.EventID, UserID: req.UserID, Seats: req.Seats}, e.Capacity)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// GET /bookings/{id}
func GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseID(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	b, err := bsvc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// GET /events/{id}/bookings
func ListBookingsByEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// ожидаем /events/{id}/bookings
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[len(parts)-1] != "bookings" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	eventID, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "container init failed")
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
		writeError(w, http.StatusInternalServerError, "failed to list bookings")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// DELETE /bookings/{id}
func CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseID(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	if err := bsvc.Cancel(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to cancel booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
