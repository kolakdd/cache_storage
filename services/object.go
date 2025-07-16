package services

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kolakdd/cache_storage/apiError"
	"github.com/kolakdd/cache_storage/models"
	"github.com/kolakdd/cache_storage/repo"
	uuid "github.com/satori/go.uuid"
)

type ObjService interface {
	UploadObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	GetObjectList(w http.ResponseWriter, r *http.Request, isHead bool) *apiError.BackendErrorInternal
	GetOneObject(w http.ResponseWriter, r *http.Request, isHead bool) *apiError.BackendErrorInternal
	DelObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
}

type objService struct {
	objRepo     repo.ObjRepo
	accessRepo  repo.AccessRepo
	queueRepo   repo.QueueRepo
	storageRepo repo.StorageRepo
	authService AuthService
	cacheRepo   repo.CacheRepo
}

func NewObjService(objRepo repo.ObjRepo, accessRepo repo.AccessRepo, amqpChan repo.QueueRepo, storageRepo repo.StorageRepo, authService AuthService, cacheRepo repo.CacheRepo) ObjService {
	return &objService{objRepo, accessRepo, amqpChan, storageRepo, authService, cacheRepo}
}

func (s *objService) UploadObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	var maxSize int64 = 500 << 20 // 500 mb
	if err := r.ParseMultipartForm(maxSize); err != nil {
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
		return apiError.InternalError
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return apiError.InternalError
	}

	// amqp logic
	err = s.queueRepo.SendUploadMessage(userID.String(), obj.ID.String())
	if err != nil {
		return err
	}

	if errTx := tx.Commit(); errTx != nil {
		return apiError.InternalError
	}

	data := models.ResponseModel{Error: nil, Response: nil, Data: models.DataResponse{obj, obj.Name}}
	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	s.cacheRepo.DelObjectList()

	return nil
}

func (s *objService) GetObjectList(w http.ResponseWriter, r *http.Request, isHead bool) *apiError.BackendErrorInternal {
	token := r.Header.Get("token")
	userID, err := s.authService.ValidateAuth(token, false)
	if err != nil {
		return err
	}

	dto := models.GetListObjectsDto{}
	dto.UserID = userID.String()
	dto.ParseValidateQuery(r.URL.Query())

	bDto, errM := json.Marshal(dto)
	if errM != nil {
		return apiError.MarshalError
	}

	cache, exist := s.cacheRepo.GetObjectList(string(bDto))
	if exist {
		headCheckWrite(w, isHead, cache)
		return nil
	}

	objList, err := s.objRepo.GetList(&dto)
	if err != nil {
		return err
	}

	data := models.ResponseModel{Error: nil, Response: nil, Data: models.DocsListResponse{&objList}}
	bData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	s.cacheRepo.SetObjectList(string(bDto), bData)
	headCheckWrite(w, isHead, bData)

	return nil
}

func (s *objService) GetOneObject(w http.ResponseWriter, r *http.Request, isHead bool) *apiError.BackendErrorInternal {
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

	cacheKey := objID.String() + ":" + fileOwnerID.String()
	cache, exist := s.cacheRepo.GetDownloadObject(cacheKey)
	if exist {
		headCheckWrite(w, isHead, cache)
		return nil
	}

	downloadURL, err := s.storageRepo.GetDownloadURL(fileOwnerID, objID, fileName)
	if err != nil {
		return err
	}

	data := models.ResponseModel{Error: nil, Response: nil, Data: models.GetDocResponse{downloadURL}}
	bData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	s.cacheRepo.SetDownloadObject(cacheKey, bData)
	headCheckWrite(w, isHead, bData)
	return nil
}

// удаляет кеш
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

	cacheKey := objID.String() + ":"
	s.cacheRepo.DelDownloadObject(cacheKey)
	jData, errM := json.Marshal(data)
	if errM != nil {
		return apiError.MarshalError
	}
	w.Write(jData)
	return nil
}

func headCheckWrite(w http.ResponseWriter, isHead bool, bData []byte) {
	if !isHead {
		w.Write(bData)
	} else {
		w.Header().Add("Content-Length", strconv.Itoa(len(bData)))
	}
}
