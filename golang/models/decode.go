package models

import (
	"encoding/json"
	"net/http"

	"github.com/kolakdd/cache_storage/golang/apiError"
)

// ValidateDecodeJSON валидирует и декудорует модель
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
