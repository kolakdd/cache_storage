package models

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/kolakdd/cache_storage/golang/apiError"
	uuid "github.com/satori/go.uuid"
)

type (
	UserXObject struct {
		UserID    uuid.UUID  `json:"id" db:"user_id"`
		ObjectID  uuid.UUID  `json:"ownerId" db:"object_id"`
		CreatedAt *time.Time `json:"createdAt" db:"created_at"`
		UpdatedAt *time.Time `json:"updatedAt" db:"updated_at"`
	}
)

type (
	DataResponse struct {
		JSON *Object `json:"json"`
		File string  `json:"file"`
	}
	DocsListResponse struct {
		Docs *[]Object `json:"docs"`
	}
	GetDocResponse struct {
		URL string `json:"url"`
	}

	Object struct {
		ID         uuid.UUID  `json:"id" db:"id"`
		OwnerID    uuid.UUID  `json:"ownerId" db:"owner_id"`
		Name       string     `json:"name" db:"name"`
		Mimetype   string     `json:"mimetype" db:"mimetype"`
		Public     bool       `json:"public" db:"public"`
		Size       int64      `json:"size" db:"size"`
		UploadS3   bool       `json:"uploadS3" db:"upload_s3"`
		IsDeleted  bool       `json:"isDeleted" db:"is_deleted"`
		Eliminated bool       `json:"eliminated" db:"eliminated"`
		CreatedAt  *time.Time `json:"createdAt" db:"created_at"`
		UpdatedAt  *time.Time `json:"updatedAt" db:"updated_at"`
	}

	UploadObjectDtoMeta struct {
		Name   string   `json:"name"`
		File   bool     `json:"file"`
		Public bool     `json:"public"`
		Token  string   `json:"token"`
		Mime   string   `json:"mime"`
		Grant  []string `json:"grant"`
	}
)

func NewObjectDB(ownerID uuid.UUID, name string, mimetype string, public bool, size int64) Object {
	now := time.Now().UTC()
	return Object{
		ID:         uuid.NewV4(),
		OwnerID:    ownerID,
		Name:       name,
		Mimetype:   mimetype,
		Public:     public,
		Size:       size,
		UploadS3:   false,
		IsDeleted:  false,
		Eliminated: false,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}
}

type GetListObjectsDto struct {
	UserID string `json:"userId"`
	Login  string `json:"login"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

func (dto *GetListObjectsDto) ParseValidateQuery(v url.Values) {
	dto.Login = v.Get("login")
	dto.Key = v.Get("key")
	dto.Value = v.Get("value")
	limitInt, err := strconv.Atoi(v.Get("limit"))
	if err != nil {
		limitInt = 50
	}
	dto.Limit = limitInt
	offsetInt, err := strconv.Atoi(v.Get("offset"))
	if err != nil {
		limitInt = 0
	}
	dto.Offset = offsetInt
}

func (dto *UploadObjectDtoMeta) ParseFormData(raw string) *apiError.BackendErrorInternal {
	err := json.Unmarshal([]byte(raw), dto)
	if err != nil {
		return apiError.BadRequest
	}
	return nil
}
