package imagecompression

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"image-compressions/compressed"
	config2 "image-compressions/internal/config"
)

func Start() error {
	cfg, err := config2.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := config2.NewLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("cannot load logger: %w", err)
	}
	logger.Info("Compressed service already running")

	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered compress image service. Error:\n", r)
		}
	}()

	compressed.ConsumerProcessing(logger, cfg)

	return nil
}
