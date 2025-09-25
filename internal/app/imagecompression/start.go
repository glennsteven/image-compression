package imagecompression

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"image-compressions/internal/compressed"
	"image-compressions/internal/config"
	"image-compressions/internal/connector"
	"image-compressions/pkg/rabbitmq"
	"image-compressions/pkg/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := config.NewLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to load logger: %w", err)
	}
	logger.Info("Compressed service starting...")

	// Setup for Graceful Shutdown
	// Create a cancelable primary context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to listen to OS signals (CTRL+C, kill)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Runs goroutine to handle shutdown signal
	go func() {
		<-shutdownChan // Block until signal is received
		logger.Info("Shutdown signal received, starting the stop process...")
		cancel() // Cancel the main context, this will signal to the customer to stop
	}()

	sessions, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Aws.Region),
		Endpoint:         aws.String(cfg.Aws.EndPoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(cfg.Aws.AccessKey, cfg.Aws.AccessSecret, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to get aws session: %w", err)
	}

	awsClient := storage.NewAwsS3(sessions)
	alerting := connector.NewAlertingDiscord(cfg.Discord, logger, http.DefaultClient)

	consumerService := compressed.NewConsumer(logger, alerting, awsClient, cfg)

	poolSize := 3
	delivery, conn, ch, err := rabbitmq.StartConsumer(ctx, cfg, poolSize)
	if err != nil {
		return fmt.Errorf("consumer rabbitmq gagal berjalan: %w", err)
	}
	defer conn.Close()
	defer ch.Close()

	consumerService.Listen(ctx, delivery)

	logger.Info("Compressed service has stopped properly.")
	return nil
}
