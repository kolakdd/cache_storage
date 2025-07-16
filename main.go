package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kolakdd/cache_storage/app"
	"github.com/kolakdd/cache_storage/database"
	"github.com/kolakdd/cache_storage/rabbitmq"
	"github.com/kolakdd/cache_storage/redis"
	"github.com/kolakdd/cache_storage/repo"
	"github.com/kolakdd/cache_storage/s3"
	"github.com/kolakdd/cache_storage/worker"
	_ "github.com/lib/pq"
)

func main() {
	initTmpDir()

	envRepo := repo.NewRepoEnv()

	db := database.InitDB(envRepo)
	defer db.Close()

	ampq := rabbitmq.InitAMQP(envRepo)
	defer ampq.Close()
	ampqCH, err := ampq.Channel()
	if err != nil {
		log.Panicf("%s: %s", "err open channel", err)
	}
	defer ampqCH.Close()

	cache := redis.InitRedis(envRepo)
	defer cache.Conn().Close()

	s3 := s3.InitS3(envRepo)

	workerChannel, _ := ampq.Channel()
	uplodWorker := worker.NewUploaderWorker(workerChannel, db, s3)
	go uplodWorker.StartConsume()
	app.App(db, cache, ampqCH, s3)
}

func initTmpDir() {
	newpath := filepath.Join(".", "tmp")
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Panic("err while create dir tmp: ", err)
	}
}
