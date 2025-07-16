package handlers

import (
	"net/http"

	"github.com/kolakdd/cache_storage/golang/apiError"
	"github.com/kolakdd/cache_storage/golang/services"
)

type ObjHandler interface {
	DocsActivity(w http.ResponseWriter, r *http.Request)
	DocsActivityID(w http.ResponseWriter, r *http.Request)
}

type objHandler struct {
	objService  services.ObjService
	authService services.AuthService
}

func NewObjectHandler(objService services.ObjService, authService services.AuthService) ObjHandler {
	return &objHandler{objService, authService}
}

func (h *objHandler) DocsActivity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodPost:
		err := h.objService.UploadObject(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	case http.MethodGet:
		err := h.objService.GetObjectList(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	default:
		apiError.BackendErrorWrite(w, apiError.MethodNotAllowed)
	}
}

func (h *objHandler) DocsActivityID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		err := h.objService.GetOneObject(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	case http.MethodDelete:
		err := h.objService.DelObject(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	default:
		apiError.BackendErrorWrite(w, apiError.MethodNotAllowed)
	}
}
