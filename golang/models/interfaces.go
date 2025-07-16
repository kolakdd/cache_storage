package models

import "github.com/kolakdd/cache_storage/golang/apiError"

type Validating interface {
	Validate() *apiError.BackendErrorInternal
}

type Cruded interface {
	CreateCRUD()
}
