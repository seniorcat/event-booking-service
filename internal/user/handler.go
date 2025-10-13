package user

import (
	"encoding/json"
	"net/http"

	"laschool.ru/event-booking-service/internal/http/handlers"
	"laschool.ru/event-booking-service/pkg/container"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handlers.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}

	urep := ctn.Get(DIUserRepo).(Repository)

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Password == "" {
		handlers.WriteError(w, http.StatusBadRequest, "password cannot be empty")
		return
	}

	id, err := urep.Create(r.Context(), &User{Email: req.Email, Name: req.Name, Password: req.Password})
	if err != nil {
		handlers.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "user registered successfully",
	})
}
