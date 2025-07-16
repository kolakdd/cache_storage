package models

import (
	"os"

	"github.com/kolakdd/cache_storage/golang/apiError"
)

type (
	RegisterDto struct {
		Token string `json:"token"`
		Login string `json:"login" validator:"min=8, regexp=^[a-zA-Z0-9]+$"`
		Pswd  string `json:"pswd" validator:"min=8"`
	}
	AuthDto struct {
		Login string `json:"login"`
		Pswd  string `json:"pswd" validator:"min=8"`
	}
	AuthResponse struct {
		Token string `json:"token"`
	}
)

func (m RegisterDto) Validate() *apiError.BackendErrorInternal {
	// todo: придумать что-то получше
	if m.Token != os.Getenv("REGISTER_TOKEN") {
		return apiError.BadToken
	}
	if err := validateLogin(m.Login); err != nil {
		return apiError.BadLogin
	}
	if err := validatePassword(m.Pswd); err != nil {
		return apiError.BadPassword
	}
	return nil
}

func (m AuthDto) Validate() *apiError.BackendErrorInternal {
	if err := validateLogin(m.Login); err != nil {
		return apiError.BadLogin
	}
	if err := validatePassword(m.Pswd); err != nil {
		return apiError.BadPassword
	}
	return nil
}

func validateLogin(login string) *apiError.BackendErrorInternal {
	return nil
}

func validatePassword(password string) *apiError.BackendErrorInternal {
	return nil
}
