package models

import (
	"encoding/json"
	"net/http"

	"github.com/kolakdd/cache_storage/apiError"
)

type Validating interface {
	Validate() *apiError.BackendErrorInternal
}

type Cruded interface {
	CreateCRUD()
}

// ValidateDecodeJSON валидирует и декодирует модель
func ValidateDecodeJSON[T Validating](r *http.Request) (*T, *apiError.BackendErrorInternal) {
	var dto T
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return nil, apiError.BadRequest
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	return &dto, nil
}
