package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	serverUtils "github.com/gadhittana01/cases-app-server/utils"
	"github.com/gadhittana01/go-modules-dependencies/utils"
	pusher "github.com/pusher/pusher-http-go/v5"
)

// NewS3Client creates a singleton S3 client
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

	// Create S3 client with custom endpoint
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.StorageEndpoint)
		o.UsePathStyle = true // Required for S3-compatible APIs
	})

	return s3Client, nil
}

// NewPresignClient creates a singleton Presign client
func NewPresignClient(s3Client *s3.Client) *s3.PresignClient {
	return s3.NewPresignClient(s3Client)
}

// NewPusherClient creates a singleton Pusher client
func NewPusherClient(config *utils.Config) *pusher.Client {
	return serverUtils.NewPusherClient(config)
}

func main() {
	config := utils.CheckAndSetConfig("./config", "app")
	DBpool, err := utils.ConnectDBPool(config.DBConnString)
	if err != nil {
		panic(err)
	}
	DB, err := utils.ConnectDB(config.DBConnString)
	if err != nil {
		panic(err)
	}

	baseConfig := &utils.BaseConfig{
		MigrationURL: config.MigrationURL,
		DBName:       config.DBName,
	}

	if err := utils.RunMigrationPool(DB, baseConfig); err != nil {
		panic(err)
	}

	// Initialize app using wire-generated code
	app, err := InitializeApp(DBpool, config)
	if err != nil {
		panic(err)
	}

	app.Start()
}
