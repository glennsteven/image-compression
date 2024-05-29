package main

import (
	"github.com/sirupsen/logrus"
	"image-compressions/cmd/image-compression/global_config"
	"image-compressions/compressed"
	"image-compressions/config"
)

func main() {
	cfg, err := config.GlobalConfig()
	if err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}

	logger := config.NewLogger(cfg)
	logger.Info("Compressed service already running")

	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered compress image service. Error:\n", r)
		}
	}()

	cfgs := global_config.Config(cfg)

	compressed.ConsumerProcessing(logger, cfgs)
}
