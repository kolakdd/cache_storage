package repo

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kolakdd/cache_storage/apiError"
	uuid "github.com/satori/go.uuid"
)

// StorageRepo отвечает за взаимодействие с объектами в хранилище (s3)
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

	internalURL, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", apiError.InternalError
	}
	u, err := url.Parse(internalURL)
	if err != nil {
		return "", apiError.InternalError
	}
	serverURL := os.Getenv("MINIO_SERVER_URL")
	if serverURL != "" && u.Host != serverURL {
		serverParsed, err := url.Parse(serverURL)
		if err != nil {
			return "", apiError.InternalError
		}
		u.Scheme = serverParsed.Scheme
		u.Host = serverParsed.Host
		newReq, _ := s.s3.GetObjectRequest(&s3.GetObjectInput{
			Bucket:                     aws.String("cache-storage-objects"),
			Key:                        aws.String(userID.String() + "/" + objID.String()),
			ResponseContentDisposition: aws.String("attachment; filename=\"" + fileName + "\""),
		})
		newReq.HTTPRequest.URL = u
		newReq.HTTPRequest.Host = serverParsed.Host

		externalURL, err := newReq.Presign(15 * time.Minute)
		if err != nil {
			return "", apiError.InternalError
		}
		return externalURL, nil
	}

	return internalURL, nil
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
