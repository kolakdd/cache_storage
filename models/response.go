package models

import apiError "github.com/kolakdd/cache_storage/apiError"

type (
	Response struct {
	}

	Data struct {
	}

	ResponseModel struct {
		Error    *apiError.BackendErrorInternal `json:"error,omitempty"`
		Response interface{}                    `json:"response,omitempty"`
		Data     interface{}                    `json:"data,omitempty"`
	}
)
