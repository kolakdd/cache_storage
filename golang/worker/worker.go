package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	AWSS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/kolakdd/cache_storage/golang/repo"

	amqp "github.com/rabbitmq/amqp091-go"
)

type UploaderWorker struct {
	ch *amqp.Channel
	db *sql.DB
	s3 *AWSS3.S3
}

func NewUploaderWorker(ampq *amqp.Channel, db *sql.DB, s3 *AWSS3.S3) UploaderWorker {
	return UploaderWorker{ch: ampq, db: db, s3: s3}
}

func (w *UploaderWorker) StartConsume() {
	q, err := w.ch.QueueDeclare(
		"upload", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)

	failOnError(err, "Failed to declare a queue")
	msgs, err := w.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	fmt.Println("WORKER: Consumer started.")
	go func() {
		for d := range msgs {
			fmt.Println("WORKER: get message")
			var data repo.UploadQueueMessage
			err := json.Unmarshal([]byte(d.Body), &data)
			if err != nil {
				fmt.Println("WORKER: error parse data. ", err)
				continue
			}
			osPath := filepath.Join("tmp", data.UserID+"."+data.FileID)
			s3Path := data.UserID + "/" + data.FileID

			putFile(osPath, s3Path, w.s3)
			err = updateFlag(w.db, data.FileID)
			if err != nil {
				fmt.Println("WORKER: Error while update flag, ", err)
			}
			err = os.Remove(osPath)
			if err != nil {
				fmt.Println("WORKER: Error while remove file, ", err)
			}
		}
	}()
}

func putFile(osPath string, s3Path string, s3 *AWSS3.S3) error {
	file, err := os.Open(osPath)
	if err != nil {
		log.Fatalf("WORKER: Failed to open file: %v", err)
	}
	defer file.Close()

	_, err = s3.PutObject(&AWSS3.PutObjectInput{
		Bucket: aws.String("cache-storage-objects"),
		Key:    aws.String(s3Path),
		Body:   file,
	})
	if err != nil {
		log.Fatalf("WORKER: UploadFile - filename: %v, err: %v", osPath, err)
	}
	return nil
}

func updateFlag(db *sql.DB, objectID string) error {
	ctx := context.Background()
	tx, _ := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})

	q := `
		UPDATE "Object" 
		SET upload_s3 = true
		WHERE id = $1
		`
	_, err := tx.Exec(q, objectID)
	if err != nil {
		return fmt.Errorf("error while exec tx, %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error while commit: %w", err)
	}
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("WORKER: %s: %s", msg, err)
	}
}
