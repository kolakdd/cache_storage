package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kolakdd/cache_storage/golang/apiError"
	amqp "github.com/rabbitmq/amqp091-go"
)

type UploadQueueMessage struct {
	UserID string `json:"userId"`
	FileID string `json:"fileId"`
}

type QueueRepo interface {
	SendUploadMessage(userID string, fileID string) *apiError.BackendErrorInternal
}

type queueRepo struct {
	channel *amqp.Channel
}

func NewQueueRepo(c *amqp.Channel) QueueRepo {
	return queueRepo{c}
}

func (r queueRepo) SendUploadMessage(userID string, fileID string) *apiError.BackendErrorInternal {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data := UploadQueueMessage{userID, fileID}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error marshal")
		return apiError.InternalError
	}
	fmt.Println("OK")

	errAMQP := r.channel.PublishWithContext(ctx,
		"",       // exchange
		"upload", // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes,
		})
	if errAMQP != nil {
		fmt.Println(errAMQP)
		return apiError.InternalError
	}
	log.Printf(" [x] Sent %s\n", data)
	return nil
}
