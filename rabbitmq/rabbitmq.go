package rabbitmq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kolakdd/cache_storage/repo"
)

func InitAMQP(envRepo repo.RepositoryEnv) *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@" + os.Getenv("RABBITMQ_HOST") + ":5672/")
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	return conn
}
