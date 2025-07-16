package apiError

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type errorResponse struct {
	Error backendError `json:"error"`
}

type backendError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type BackendErrorInternal struct {
	Code     int    `json:"code"`
	Text     string `json:"text"`
	HTTPCode int
}

func (e BackendErrorInternal) Error() string {
	return fmt.Sprint("code: %s, text: %s, httpCode: %s \n", e.Code, e.Text, e.HTTPCode)
}

var (
	BadRequest       = &BackendErrorInternal{0, "Неверная форма запроса", http.StatusBadRequest}
	NotFound         = &BackendErrorInternal{33, "Не найдено", http.StatusNotFound}
	MethodNotAllowed = &BackendErrorInternal{44, "Метод запрещен", http.StatusMethodNotAllowed}
	InternalError    = &BackendErrorInternal{55, "Внутренняя ошибка", http.StatusInternalServerError}
	Unauthorized     = &BackendErrorInternal{66, "Unauthorized", http.StatusUnauthorized}
)

var (
	// auth error
	BadToken    = &BackendErrorInternal{100, "Неверный токен авторизации", http.StatusMethodNotAllowed}
	BadLogin    = &BackendErrorInternal{101, "Невалидный логин", http.StatusBadRequest}
	BadPassword = &BackendErrorInternal{102, "Невалидный пароль", http.StatusBadRequest}
	// user error
	UserAlreadyExist = &BackendErrorInternal{103, "Пользователь уже существует", http.StatusForbidden}
	MarshalError     = &BackendErrorInternal{104, "Ошибка сериализации", http.StatusInternalServerError}
	RedisError       = &BackendErrorInternal{105, "Ошибка при работе с Redis", http.StatusInternalServerError}
	// file
	FileToBig    = &BackendErrorInternal{106, "Загружаемый файл слишком велик", http.StatusBadRequest}
	FileGetError = &BackendErrorInternal{107, "Ошибка при получении файла", http.StatusBadRequest}

	InvalidObjectKey = &BackendErrorInternal{108, "Неверный ключ", http.StatusBadRequest}
	// storage
	StorageGetError    = &BackendErrorInternal{109, "Ошибка получения файла", http.StatusInternalServerError}
	StorageDeleteError = &BackendErrorInternal{110, "Ошибка удаления файла", http.StatusInternalServerError}
)

func BackendErrorWrite(w http.ResponseWriter, err *BackendErrorInternal) {
	w.WriteHeader(err.HTTPCode)
	jData, _ := json.Marshal(errorResponse{Error: backendError{Code: err.Code, Text: err.Text}})
	w.Write(jData)
}
