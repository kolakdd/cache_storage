package handlers

import (
	"net/http"

	"github.com/kolakdd/cache_storage/apiError"
	"github.com/kolakdd/cache_storage/services"
	"github.com/kolakdd/cache_storage/slogger"
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
	slogger.LoggerHandler(r)
	meth := r.Method
	switch meth {
	case http.MethodPost:
		err := h.objService.UploadObject(w, r)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	case http.MethodGet, http.MethodHead:
		var isHead bool
		if meth == http.MethodHead {
			isHead = true
		} else {
			isHead = false
		}
		err := h.objService.GetObjectList(w, r, isHead)
		if err != nil {
			apiError.BackendErrorWrite(w, err)
		}
	default:
		apiError.BackendErrorWrite(w, apiError.MethodNotAllowed)
	}
}

func (h *objHandler) DocsActivityID(w http.ResponseWriter, r *http.Request) {
	slogger.LoggerHandler(r)
	w.Header().Set("Content-Type", "application/json")
	meth := r.Method
	switch meth {
	case http.MethodGet, http.MethodHead:
		var isHead bool
		if meth == http.MethodHead {
			isHead = true
		} else {
			isHead = false
		}
		err := h.objService.GetOneObject(w, r, isHead)
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
