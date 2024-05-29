package imagecompression

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"image-compressions/internal/compressed"
	"image-compressions/internal/config"
	"image-compressions/pkg/rabbitmq"
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

	delivery, conn, err := rabbitmq.Consumer(cfg.RabbitMq)
	if err != nil {
		return fmt.Errorf("consumer rabbit failed running: %w", err)
	}

	defer func() {
		conn.Close()
	}()

	consumer := compressed.NewConsumer(logger, delivery)
	consumer.Listen(cfg)

	return nil
}
