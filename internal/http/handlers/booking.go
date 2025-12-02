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
// CreateBooking godoc
// @Summary      Создать бронирование
// @Description  Создает новое бронирование для события
// @Tags         bookings
// @Security     Bearer
// @Accept       json
// @Produce      json
// @Param        booking  body  booking.CreateBookingRequest  true  "Данные бронирования"
// @Success      201  {object}  map[string]int64  "id of created booking"
// @Failure      400  {object}  handlers.ErrorResponse
// @Router       /bookings [post]
func CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	esvc := ctn.Get(event.DIEventService).(event.Service)

	var req booking.CreateBookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	e, err := esvc.Get(r.Context(), req.EventID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "event not found")
		return
	}

	id, err := bsvc.Create(r.Context(), &booking.Booking{EventID: req.EventID, UserID: req.UserID, Seats: req.Seats}, e.Capacity)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// GetBooking godoc
// @Summary      Получить бронирование
// @Description  Возвращает информацию о бронировании по ID
// @Tags         bookings
// @Produce      json
// @Param        id   path      int  true  "ID бронирования"
// @Success      200  {object}  booking.Booking  "Пример успешного ответа"
// @Failure      400  {object}  handlers.ErrorResponse  "Некорректный ID"
// @Failure      404  {object}  handlers.ErrorResponse  "Бронирование не найдено"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /bookings/{id} [get]
func GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseID(r.URL.Path)
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	b, err := bsvc.Get(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// ListBookingsByEvent godoc
// @Summary      Список бронирований по событию
// @Description  Возвращает список бронирований для события
// @Tags         bookings
// @Produce      json
// @Param        id      path   int  true  "ID события"
// @Param        limit   query  int  false "Лимит записей"
// @Param        offset  query  int  false "Смещение"
// @Success      200  {array}  booking.Booking  "Пример успешного ответа"
// @Failure      400  {object}  handlers.ErrorResponse  "Некорректный ID события"
// @Failure      404  {object}  handlers.ErrorResponse  "Событие не найдено"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /events/{id}/bookings [get]
func ListBookingsByEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// ожидаем /events/{id}/bookings
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[len(parts)-1] != "bookings" {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	eventID, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
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
		WriteError(w, http.StatusInternalServerError, "failed to list bookings")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// CancelBooking godoc
// @Summary      Отменить бронирование
// @Description  Отменяет бронирование по ID
// @Tags         bookings
// @Security     Bearer
// @Param        id   path      int  true  "ID бронирования"
// @Success      204  "Бронирование отменено"
// @Failure      400  {object}  handlers.ErrorResponse  "Некорректный ID"
// @Failure      404  {object}  handlers.ErrorResponse  "Бронирование не найдено"
// @Failure      500  {object}  handlers.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /bookings/{id} [delete]
func CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := parseID(r.URL.Path)
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	if err := bsvc.Cancel(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to cancel booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
