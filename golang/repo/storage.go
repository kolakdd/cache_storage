package repo

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kolakdd/cache_storage/golang/apiError"
	uuid "github.com/satori/go.uuid"
)

type StorageRepo interface {
	GetDownloadURL(userID uuid.UUID, objID uuid.UUID, fileName string) (string, *apiError.BackendErrorInternal)
	Delete(ownerID uuid.UUID, objID uuid.UUID) *apiError.BackendErrorInternal
}

type storageRepo struct {
	s3 *s3.S3
}

func NewStorageRepo(storage *s3.S3) StorageRepo {
	return &storageRepo{storage}
}

func (s *storageRepo) GetDownloadURL(userID uuid.UUID, objID uuid.UUID, fileName string) (string, *apiError.BackendErrorInternal) {
	req, _ := s.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String("cache-storage-objects"),
		Key:                        aws.String(userID.String() + "/" + objID.String()),
		ResponseContentDisposition: aws.String("attachment; filename=\"" + fileName + "\""),
	})
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {

		log.Println("Failed to sign request", err)
	}
	return urlStr, nil
}

func (s *storageRepo) Delete(ownerID uuid.UUID, objID uuid.UUID) *apiError.BackendErrorInternal {
	_, err := s.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String("cache-storage-objects"),
		Key:    aws.String(ownerID.String() + "/" + objID.String()),
	})
	if err != nil {
		log.Println("Failed to sign request", err)
		return apiError.StorageDeleteError

	}
	return nil
}
