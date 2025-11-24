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

// CreateEvent godoc
// @Summary      Создать событие
// @Description  Создает новое событие
// @Tags         events
// @Security     Bearer
// @Accept       json
// @Produce      json
// @Param        event  body  event.CreateEventRequest  true  "Данные события"
// @Success      201  {object}  event.Event
// @Failure      400  {object}  handlers.ErrorResponse
// @Router       /events [post]
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

// GetEvent godoc
// @Summary      Получить событие
// @Description  Возвращает информацию о событии по ID
// @Tags         events
// @Produce      json
// @Param        id   path      int  true  "ID события"
// @Success      200  {object}  event.Event  "Пример успешного ответа"
// @Failure      400  {object}  handlers.ErrorResponse  "Некорректный ID"
// @Failure      404  {object}  handlers.ErrorResponse  "Событие не найдено"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /events/{id} [get]
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

// ListEvents godoc
// @Summary      Список событий
// @Description  Возвращает список событий
// @Tags         events
// @Produce      json
// @Param        limit   query  int  false "Лимит записей"
// @Param        offset  query  int  false "Смещение"
// @Success      200  {array}  event.Event  "Пример успешного ответа"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /events [get]
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

// UpdateEvent godoc
// @Summary      Обновить событие
// @Description  Обновляет данные события по ID
// @Tags         events
// @Security     Bearer
// @Accept       json
// @Param        id     path   int  true  "ID события"
// @Param        event  body   event.CreateEventRequest  true  "Данные события"
// @Success      204  "Событие обновлено"
// @Failure      400  {object}  handlers.ErrorResponse  "Некорректные данные"
// @Failure      404  {object}  handlers.ErrorResponse  "Событие не найдено"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /events/{id} [put]
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

// DeleteEvent godoc
// @Summary      Удалить событие
// @Description  Удаляет событие по ID
// @Tags         events
// @Security     Bearer
// @Param        id   path      int  true  "ID события"
// @Success      204  "Событие удалено"
// @Failure      400  {object}  handlers.ErrorResponse  "Некорректный ID"
// @Failure      404  {object}  handlers.ErrorResponse  "Событие не найдено"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /events/{id} [delete]
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
