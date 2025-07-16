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
)

type ObjService interface {
	UploadObject(w http.ResponseWriter, r *http.Request, authServ AuthService) *apiError.BackendErrorInternal
	GetObjectList(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	GetOneObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
	DelObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal
}

type objService struct {
	objRepo    repo.ObjRepo
	accessRepo repo.AccessRepo
	queueRepo  repo.QueueRepo
}

func NewObjService(objRepo repo.ObjRepo, accessRepo repo.AccessRepo, amqpChan repo.QueueRepo) ObjService {
	return &objService{objRepo, accessRepo, amqpChan}
}

func (s *objService) UploadObject(w http.ResponseWriter, r *http.Request, authS AuthService) *apiError.BackendErrorInternal {
	var maxSize int64 = 10 << 20 // 10mb
	if err := r.ParseMultipartForm(maxSize); err != nil {
		fmt.Println(err)
		return apiError.FileToBig
	}

	var dto models.UploadObjectDtoMeta
	if err := dto.ParseFormData(r.FormValue("meta")); err != nil {
		return err
	}

	userID, err := authS.ValidateAuth(dto.Token, false)
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

	err = s.queueRepo.SendUploadMessage(userID.String(), obj.ID.String())
	if err != nil {
		return err
	}
	// amqp logic
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
	return nil
}

func (s *objService) GetOneObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	return nil

}

func (s *objService) DelObject(w http.ResponseWriter, r *http.Request) *apiError.BackendErrorInternal {
	return nil

}
