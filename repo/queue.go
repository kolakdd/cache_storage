package repo

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/kolakdd/cache_storage/apiError"
	amqp "github.com/rabbitmq/amqp091-go"
)

type UploadQueueMessage struct {
	UserID string `json:"userId"`
	FileID string `json:"fileId"`
}

// QueueRepo взаимодействие с брокером. В частности за отравку
// сообщений в очередь, для дальнейшей асинхронной загрузки в хранилище
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
		return apiError.InternalError
	}
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
		return apiError.InternalError
	}
	log.Printf(" [x] Sent %s\n", data)
	return nil
}
