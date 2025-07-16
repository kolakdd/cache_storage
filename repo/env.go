package repo

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// RepositoryEnv отвечает за хранение и получение переменных окружения
type RepositoryEnv interface {
	GetDatabaseDSN() string
	GetAPIMode() string
	GetSecret() string
	GetRegisterToken() string
	GetS3Cred() (string, string)
	GetUploadBucket() string
	GetRegisURL() string
}

type repositoryEnv struct {
	pgUser     string `env:"POSTGRES_USER"`
	pgPassword string `env:"POSTGRES_PASSWORD"`
	pgHost     string `env:"PG_HOST"`
	dbName     string `env:"DB_NAME"`
	dbPort     int    `env:"DB_PORT"`

	registerToken string `env:"REGISTER_TOKEN"`
	apiSecret     string `env:"API_SECRET"`

	s3User      string `env:"MINIO_ROOT_USER"`
	s3Pass      string `env:"MINIO_ROOT_PASSWORD"`
	s3Url       string `env:"MINIO_URL"`
	s3UrlServer string `env:"MINIO_SERVER_URL"`

	s3UploadBucket string `env:"UPLOAD_MAIN_BUCKET"`

	redisURL     string `env:"REDIS_URL"`
	rabbitmqHost string `env:"RABBITMQ_HOST"`
	mode         string `env:"MODE"`
}

func NewRepoEnv() RepositoryEnv {
	if err := godotenv.Load(); err != nil {
		log.Panic("No .env file found: ", err)
	}
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgHost := os.Getenv("PG_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := parseInt("DB_PORT")

	registerToken := os.Getenv("REGISTER_TOKEN")
	apiSecret := os.Getenv("API_SECRET")

	s3User := os.Getenv("MINIO_ROOT_USER")
	s3Pass := os.Getenv("MINIO_ROOT_PASSWORD")
	s3Url := os.Getenv("MINIO_URL")
	s3UrlServer := os.Getenv("MINIO_SERVER_URL")

	s3UploadBucket := os.Getenv("UPLOAD_MAIN_BUCKET")

	redisURL := os.Getenv("REDIS_URL")
	rabbitmqHost := os.Getenv("RABBITMQ_HOST")

	mode := os.Getenv("MODE")

	return &repositoryEnv{pgUser, pgPassword, pgHost, dbName, dbPort, registerToken, apiSecret, s3User, s3Pass, s3Url, s3UrlServer, s3UploadBucket, redisURL, rabbitmqHost, mode}
}

func parseInt(evnKey string) int {
	v := os.Getenv(evnKey)
	vInt, err := strconv.Atoi(v)
	if err != nil {
		log.Panicf("err while parse env key=%s", evnKey)
	}
	return vInt

}

func (r *repositoryEnv) GetDatabaseDSN() string {
	user := r.pgUser
	password := r.pgPassword
	host := r.pgHost
	db := r.dbName
	port := r.dbPort
	return fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%d sslmode=disable", user, password, host, db, port)
}

func (r *repositoryEnv) GetAPIMode() string {
	return r.mode
}

func (r *repositoryEnv) GetSecret() string {
	return r.apiSecret
}

func (r *repositoryEnv) GetRegisterToken() string {
	return r.registerToken
}

func (r *repositoryEnv) GetS3Cred() (string, string) {
	return r.s3User, r.s3Pass
}

func (r *repositoryEnv) GetUploadBucket() string {
	return r.s3UploadBucket
}

func (r *repositoryEnv) GetRegisURL() string {
	return r.redisURL
}
