package s3

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kolakdd/cache_storage/repo"
)

func InitS3(env repo.RepositoryEnv) *s3.S3 {
	conf := aws.NewConfig()
	conf.Endpoint = aws.String(os.Getenv("MINIO_URL"))
	conf.DisableSSL = aws.Bool(true)
	conf.S3ForcePathStyle = aws.Bool(true) // для minio
	conf.Region = aws.String("us-east-1")
	user, pass := env.GetS3Cred()
	conf.Credentials = credentials.NewStaticCredentials(user, pass, "")

	sess, _ := session.NewSession(conf)
	svc := s3.New(sess)

	if err := initBuckets(env, svc); err != nil {
		panic(err)
	}
	return svc
}

func initBuckets(env repo.RepositoryEnv, svc *s3.S3) error {
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(env.GetUploadBucket()),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			default:
				return err
			}
		}
	}
	return nil
}
