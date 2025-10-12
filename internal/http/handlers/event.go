package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"laschool.ru/event-booking-service/internal/event"
	"laschool.ru/event-booking-service/pkg/container"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseIDFromPath(path string) (int64, bool) {
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

// POST /events
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	svc := ctn.Get(event.DIEventService).(event.Service)

	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Location    string    `json:"location"`
		StartsAt    time.Time `json:"starts_at"`
		EndsAt      time.Time `json:"ends_at"`
		Capacity    int       `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	id, err := svc.Create(r.Context(), &event.Event{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Capacity:    req.Capacity,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// GET /events/{id}
func GetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	svc := ctn.Get(event.DIEventService).(event.Service)

	e, err := svc.Get(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// GET /events
func ListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	svc := ctn.Get(event.DIEventService).(event.Service)

	// простые параметры пагинации из query
	limit := 20
	offset := 0
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

	list, err := svc.List(r.Context(), limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// PUT /events/{id}
func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	svc := ctn.Get(event.DIEventService).(event.Service)

	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Location    string    `json:"location"`
		StartsAt    time.Time `json:"starts_at"`
		EndsAt      time.Time `json:"ends_at"`
		Capacity    int       `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	err = svc.Update(r.Context(), &event.Event{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Capacity:    req.Capacity,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /events/{id}
func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	svc := ctn.Get(event.DIEventService).(event.Service)

	if err := svc.Delete(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
