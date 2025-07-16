package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kolakdd/cache_storage/golang/repo"
)

func InitAMQP(envRepo repo.RepositoryEnv) *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	return conn
}
