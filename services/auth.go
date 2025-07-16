package services

import (
	"encoding/json"
	"net/http"

	"github.com/kolakdd/cache_storage/apiError"
	"github.com/kolakdd/cache_storage/models"
	"github.com/kolakdd/cache_storage/repo"
	uuid "github.com/satori/go.uuid"
)

type AuthService interface {
	RegisterUser(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	AuthUser(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	Unlogin(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	ValidateAuth(token string, del bool) (*uuid.UUID, *apiError.BackendErrorInternal)
}

type authService struct {
	authRepo repo.AuthRepo
}

func NewAuthService(authRepo repo.AuthRepo) AuthService {
	return &authService{authRepo}
}

func (s *authService) RegisterUser(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	dto, err := models.ValidateDecodeJSON[models.RegisterDto](r)
	if err != nil {
		return err
	}
	res, err := s.authRepo.CreateUser(dto)
	if err != nil {
		return err
	}
	jData, errM := json.Marshal(res)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	return nil
}

func (s *authService) AuthUser(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	var err *apiError.BackendErrorInternal

	dto, err := models.ValidateDecodeJSON[models.AuthDto](r)
	if err != nil {
		return err
	}
	id, err := s.authRepo.GetUserID(dto)
	if err != nil {
		return err
	}
	token, err := s.authRepo.CreateAuthToken(id)
	if err != nil {
		return err
	}
	data := models.ResponseModel{Error: nil, Response: models.AuthResponse{Token: *token}, Data: nil}

	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	return nil
}

func (s *authService) Unlogin(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	var err *apiError.BackendErrorInternal
	tokenStr := r.PathValue("token")

	_, err = s.ValidateAuth(tokenStr, true)
	if err != nil {
		return err
	}
	resp := make(map[string]bool)
	resp[tokenStr] = true

	data := models.ResponseModel{Error: nil, Response: resp, Data: nil}

	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	return nil
}

func (s *authService) ValidateAuth(token string, del bool) (*uuid.UUID, *apiError.BackendErrorInternal) {
	id, err := s.authRepo.GetByTokenAndUnauthorize(token, del)
	if err != nil {
		return nil, err
	}
	return id, nil
}
