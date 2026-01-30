package providers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gadhittana01/cases-modules/utils"
	pusher "github.com/pusher/pusher-http-go/v5"
)


func NewS3Client(config *utils.Config) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.StorageRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.StorageAccessKey,
			config.StorageSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}


	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.StorageEndpoint)
		o.UsePathStyle = true
	})

	return s3Client, nil
}


func NewPresignClient(s3Client *s3.Client) *s3.PresignClient {
	return s3.NewPresignClient(s3Client)
}


func NewPusherClient(config *utils.Config) *pusher.Client {
	return utils.NewPusherClient(config)
}
