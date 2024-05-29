package imagecompression

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"image-compressions/compressed"
	"image-compressions/config"
)

func Start() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := config.NewLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	logger.Info("Compressed service already running")
	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered compress image service. Error:\n", r)
		}
	}()

	compressed.ConsumerProcessing(logger, cfg.RabbitMQ, cfg.Discord, cfg.Server, cfg.ImageSetting)
	return nil
}
