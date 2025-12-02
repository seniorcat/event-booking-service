package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"laschool.ru/event-booking-service/internal/booking"
	"laschool.ru/event-booking-service/internal/cache"
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
	cachesrv := ctn.Get(cache.DICacheService).(cache.Service)

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

	newBooking := &booking.Booking{
		EventID:   req.EventID,
		UserID:    req.UserID,
		Seats:     req.Seats,
		Status:    "confirmed",
		CreatedAt: time.Now(),
	}
	id, err := bsvc.Create(r.Context(), newBooking, e.Capacity)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Удаляем кэш для этого события
		pattern := fmt.Sprintf("event:%d:bookings*", req.EventID)
		if err := cachesrv.DeletePattern(ctx, pattern); err != nil {
			log.Printf("WARNING: Cache invalidation failed for event %d: %v", req.EventID, err)
		} else {
			log.Printf("Cache invalidated for event %d after booking creation", req.EventID)
		}

		bookingKey := fmt.Sprintf("booking:%d", id)
		if err := cachesrv.Set(ctx, bookingKey, newBooking, 30*time.Minute); err != nil {
			log.Printf("WARNING: Failed to cache booking %d: %v", id, err)
		} else {
			log.Printf("Booking %d cached successfully", id)
		}
	}()
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
	cacheService := ctn.Get(cache.DICacheService).(cache.Service)

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

	if r.URL.Query().Get("refresh") == "true" {
		// Удаляем ТОЛЬКО бронирования этого события
		pattern := fmt.Sprintf("event:%d:bookings:*", eventID)
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := cacheService.DeletePattern(ctx, pattern); err != nil {
			log.Printf("Cache invalidation failed for event %d: %v", eventID, err)
			// Не возвращаем ошибку клиенту - продолжаем работу
		} else {
			log.Printf("Cache invalidated for event %d bookings (manual refresh)", eventID)
		}
	}

	cacheKey := fmt.Sprintf("event:%d:bookings:limit:%d:offset:%d", eventID, limit, offset)

	calculateFunc := func() (interface{}, error) {
		return bsvc.ListByEvent(r.Context(), eventID, limit, offset)
	}

	data, err := cacheService.GetProtected(r.Context(), cacheKey, calculateFunc, 5*time.Minute)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list bookings")
		return
	}
	var bookings []booking.Booking
	if err := json.Unmarshal(data, &bookings); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to parse cached data")
		return
	}
	writeJSON(w, http.StatusOK, bookings)
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
	cacheService := ctn.Get(cache.DICacheService).(cache.Service)

	if err := bsvc.Cancel(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to cancel booking")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Удаляем все связанное с бронированиями
		cacheService.DeletePattern(ctx, "event:*:bookings*")
		cacheService.DeletePattern(ctx, fmt.Sprintf("booking:%d", id))
		cacheService.DeletePattern(ctx, "stats:bookings*")
	}()
	w.WriteHeader(http.StatusNoContent)
}
