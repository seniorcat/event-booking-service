package user

import (
	"encoding/json"
	"net/http"

	"laschool.ru/event-booking-service/internal/http/handlers"
	"laschool.ru/event-booking-service/pkg/container"
)

// RegisterHandler godoc
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body  user.RegisterRequest  true  "Данные пользователя"
// @Success      201  {object}  map[string]interface{} "id"
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      409  {object}  handlers.ErrorResponse
// @Router       /users/register [post]
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

	userv := ctn.Get(DIUserService).(Service)

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Password == "" {
		handlers.WriteError(w, http.StatusBadRequest, "password cannot be empty")
		return
	}

	id, err := userv.Register(r.Context(), &User{Email: req.Email, Name: req.Name, Password: req.Password})
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

// LoginHandler godoc
// @Summary      Вход в систему
// @Description  Выполняет аутентификацию и возвращает JWT
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body  user.LoginRequest  true  "Email и пароль"
// @Success      200  {object}  map[string]interface{} "id и token"
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      401  {object}  handlers.ErrorResponse
// @Router       /users/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handlers.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, "container init failed")
		return
	}

	userv := ctn.Get(DIUserService).(Service)

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Password == "" {
		handlers.WriteError(w, http.StatusUnauthorized, "password cannot be empty")
		return
	}
	if len(req.Email) < 6 {
		handlers.WriteError(w, http.StatusUnauthorized, "email is too small")
		return
	}

	token, err := userv.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handlers.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
	})

}
