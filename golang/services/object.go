package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kolakdd/cache_storage/golang/apiError"
	"github.com/kolakdd/cache_storage/golang/models"
	"github.com/kolakdd/cache_storage/golang/repo"
	uuid "github.com/satori/go.uuid"
)

type ObjService interface {
	UploadObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	GetObjectList(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	GetOneObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	DelObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
}

type objService struct {
	objRepo     repo.ObjRepo
	accessRepo  repo.AccessRepo
	queueRepo   repo.QueueRepo
	storageRepo repo.StorageRepo
	authService AuthService
}

func NewObjService(objRepo repo.ObjRepo, accessRepo repo.AccessRepo, amqpChan repo.QueueRepo, storageRepo repo.StorageRepo, authService AuthService) ObjService {
	return &objService{objRepo, accessRepo, amqpChan, storageRepo, authService}
}

func (s *objService) UploadObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	var maxSize int64 = 500 << 20 // 500 mb
	if err := r.ParseMultipartForm(maxSize); err != nil {
		fmt.Println(err)
		return apiError.FileToBig
	}

	var dto models.UploadObjectDtoMeta
	if err := dto.ParseFormData(r.FormValue("meta")); err != nil {
		return err
	}

	userID, err := s.authService.ValidateAuth(dto.Token, false)
	if err != nil {
		return err
	}

	file, handler, errF := r.FormFile("file")
	if errF != nil {
		fmt.Println("error while get file", errF)
		return apiError.FileGetError
	}
	defer file.Close()

	tx := s.objRepo.CreateTX()

	obj, err := s.objRepo.Create(tx, &dto, *userID, handler.Size)
	if err != nil {
		return apiError.FileGetError
	}
	err = s.accessRepo.CreateMany(tx, obj.ID, dto.Grant)
	if err != nil {
		return apiError.FileGetError
	}
	// файлы сохраняются во временную папку tmp с именем {userID}.{objID}
	dst, errF := os.Create(filepath.Join("tmp", userID.String()+"."+obj.ID.String()))
	if errF != nil {
		fmt.Println("error while create file", err)
		return apiError.InternalError
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		fmt.Println("error while copy file", err)
		return apiError.InternalError
	}

	// amqp logic
	err = s.queueRepo.SendUploadMessage(userID.String(), obj.ID.String())
	if err != nil {
		return err
	}

	if errTx := tx.Commit(); errTx != nil {
		fmt.Println("failed to commit transaction: %w", errTx)
		return apiError.InternalError
	}

	data := models.ResponseModel{Error: nil, Response: nil, Data: models.DataResponse{obj, obj.Name}}
	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)

	return nil
}

func (s *objService) GetObjectList(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	token := r.Header.Get("token")
	userID, err := s.authService.ValidateAuth(token, false)
	if err != nil {
		return err
	}
	dto := models.GetListObjectsDto{}
	dto.UserID = userID.String()
	dto.ParseValidateQuery(r.URL.Query())

	fmt.Println(dto)
	objList, err := s.objRepo.GetList(&dto)
	if err != nil {
		return err
	}

	data := models.ResponseModel{Error: nil, Response: nil, Data: models.DocsListResponse{&objList}}
	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)

	return nil
}

func (s *objService) GetOneObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	token := r.Header.Get("token")
	userID, err := s.authService.ValidateAuth(token, false)
	if err != nil {
		return err
	}

	objID, pErr := uuid.FromString(r.PathValue("id"))
	if pErr != nil {
		return apiError.BadRequest
	}

	fileOwnerID, fileName, err := s.objRepo.CheckAccess(objID, *userID)
	if err != nil {
		return err
	}
	downloadURL, err := s.storageRepo.GetDownloadURL(fileOwnerID, objID, fileName)
	if err != nil {
		return err
	}

	data := models.ResponseModel{Error: nil, Response: nil, Data: models.GetDocResponse{downloadURL}}
	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	return nil
}

func (s *objService) DelObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	token := r.Header.Get("token")
	userID, err := s.authService.ValidateAuth(token, false)
	if err != nil {
		return err
	}

	objID, pErr := uuid.FromString(r.PathValue("id"))
	if pErr != nil {
		return apiError.BadRequest
	}

	fileOwnerID, _, err := s.objRepo.CheckAccess(objID, *userID)
	if err != nil {
		return err
	}

	err = s.objRepo.DeleteWithAccess(objID)
	if err != nil {
		return err
	}

	err = s.storageRepo.Delete(fileOwnerID, objID)
	if err != nil {
		return err
	}

	resp := make(map[string]bool)
	resp[objID.String()] = true
	data := models.ResponseModel{Error: nil, Response: resp, Data: nil}

	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	return nil
}
