package imagecompression

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"image-compressions/internal/compressed"
	"image-compressions/internal/config"
	"image-compressions/internal/connector"
	"image-compressions/pkg/rabbitmq"
	"image-compressions/pkg/storage"
	"net/http"
)

func Start() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := config.NewLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("cannot load logger: %w", err)
	}
	logger.Info("Compressed service already running")

	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered compress image service. Error:\n", r)
		}
	}()

	ctx := context.Background()
	delivery, conn, err := rabbitmq.Consumer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("consumer rabbit failed running: %w", err)
	}

	defer func() {
		conn.Close()
	}()

	sessions, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Aws.Region),
		Endpoint:         aws.String(cfg.Aws.EndPoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			cfg.Aws.AccessKey,
			cfg.Aws.AccessSecret,
			"",
		),
	})

	if err != nil {
		logrus.Println("NewSession. Error:\n", err)
		return fmt.Errorf("failed to get session aws: %w", err)
	}

	awsClient := storage.NewAwsS3(sessions)
	alerting := connector.NewAlertingDiscord(cfg.Discord, logger, http.DefaultClient)
	consumer := compressed.NewConsumer(logger, delivery, alerting, awsClient)
	consumer.Listen(cfg)

	return nil
}
