package handlers

import (
	"net/http"

	"github.com/kolakdd/cache_storage/apiError"
	"github.com/kolakdd/cache_storage/services"
	"github.com/kolakdd/cache_storage/slogger"
)

type AuthHandler interface {
	RegisterUserHandler(w http.ResponseWriter, r *http.Request)
	LoginUserHandler(w http.ResponseWriter, r *http.Request)
	UnloginUserHandler(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) AuthHandler {
	return &authHandler{authService}
}

func (h *authHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	slogger.LoggerHandler(r)
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodPost:
		err := h.authService.RegisterUser(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	default:
		apiError.BackendErrorWrite(w, apiError.MethodNotAllowed)
	}
}

func (h *authHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	slogger.LoggerHandler(r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		apiError.BackendErrorWrite(w, apiError.MethodNotAllowed)
		return
	}
	err := h.authService.AuthUser(w, r)
	if err != nil {
		apiError.BackendErrorWrite(w, err)
	}
}

func (h *authHandler) UnloginUserHandler(w http.ResponseWriter, r *http.Request) {
	slogger.LoggerHandler(r)
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodDelete:
		err := h.authService.Unlogin(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	default:
		apiError.BackendErrorWrite(w, apiError.MethodNotAllowed)
	}
}
