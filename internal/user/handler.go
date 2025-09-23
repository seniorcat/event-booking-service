package user

import "net/http"

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: добавить логику регистрации пользователя
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user registered"))
}
