package models

import (
	"os"
	"regexp"

	"github.com/kolakdd/cache_storage/apiError"
)

type (
	RegisterDto struct {
		Token string `json:"token"`
		Login string `json:"login""`
		Pswd  string `json:"pswd"`
	}
	AuthDto struct {
		Login string `json:"login"`
		Pswd  string `json:"pswd"`
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
		return err
	}
	if err := validatePassword(m.Pswd); err != nil {
		return err
	}
	return nil
}

func (m AuthDto) Validate() *apiError.BackendErrorInternal {
	if err := validateLogin(m.Login); err != nil {
		return err
	}
	if err := validatePassword(m.Pswd); err != nil {
		return err
	}
	return nil
}

func validateLogin(login string) *apiError.BackendErrorInternal {
	if len(login) < 8 {
		return apiError.LoginToEasy
	}
	validLoginRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !validLoginRegex.MatchString(login) {
		return apiError.LoginToEasy
	}
	return nil
}

func validatePassword(password string) *apiError.BackendErrorInternal {
	if len(password) < 8 {
		return apiError.PasswordToEasy
	}
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return apiError.PasswordToEasy
	}
	hasSpecialChar := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	if !hasSpecialChar {
		return apiError.PasswordToEasy
	}
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasLower || !hasUpper {
		return apiError.PasswordToEasy
	}
	return nil
}
