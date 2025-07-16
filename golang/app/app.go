package app

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kolakdd/cache_storage/golang/handlers"
	"github.com/kolakdd/cache_storage/golang/repo"
	"github.com/kolakdd/cache_storage/golang/services"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func App(db *sql.DB, cache *redis.Client, amqp *amqp.Channel, s3 *s3.S3) {
	authRepo := repo.NewAuthRepo(db, cache)
	objRepo := repo.NewObjRepo(db)
	accessRepo := repo.NewAccessRepo(db)

	queueRepo := repo.NewQueueRepo(amqp)

	authService := services.NewAuthService(authRepo)
	objService := services.NewObjService(objRepo, accessRepo, queueRepo)

	mux := http.NewServeMux()

	authHandler := handlers.NewAuthHandler(authService)

	mux.HandleFunc("/api/register", authHandler.RegisterUserHandler)
	mux.HandleFunc("/api/login", authHandler.LoginUserHandler)
	mux.HandleFunc("/api/login/{token}", authHandler.UnloginUserHandler)

	objHandler := handlers.NewObjectHandler(objService, authService)

	mux.HandleFunc("/api/docs", objHandler.DocsActivity)
	mux.HandleFunc("/api/docs/{id}", objHandler.DocsActivityToken)

	err := http.ListenAndServe(":8090", mux)
	log.Fatal(err)
}
