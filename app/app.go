package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kolakdd/cache_storage/handlers"
	"github.com/kolakdd/cache_storage/repo"
	"github.com/kolakdd/cache_storage/services"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func App(db *sql.DB, redis *redis.Client, amqp *amqp.Channel, s3 *s3.S3) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	authRepo := repo.NewAuthRepo(db, redis)
	objRepo := repo.NewObjRepo(db)
	accessRepo := repo.NewAccessRepo(db)

	storageRepo := repo.NewStorageRepo(s3)
	queueRepo := repo.NewQueueRepo(amqp)
	cacheRepo := repo.NewCacheRepo(redis)

	authService := services.NewAuthService(authRepo)
	objService := services.NewObjService(objRepo, accessRepo, queueRepo, storageRepo, authService, cacheRepo)

	mux := http.NewServeMux()

	authHandler := handlers.NewAuthHandler(authService)

	mux.HandleFunc("/api/register", authHandler.RegisterUserHandler)
	mux.HandleFunc("/api/login", authHandler.LoginUserHandler)
	mux.HandleFunc("/api/login/{token}", authHandler.UnloginUserHandler)

	objHandler := handlers.NewObjectHandler(objService, authService)

	mux.HandleFunc("/api/docs", objHandler.DocsActivity)
	mux.HandleFunc("/api/docs/{id}", objHandler.DocsActivityID)

	logger.Println("Server is starting...")
	err := http.ListenAndServe(":8090", mux)
	if err != nil {
		logger.Fatal("ListenAndServe: ", err)
	}
}
