package main

import (
	"github.com/sirupsen/logrus"
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

	rabbitCfg := config.RabbitMq{
		Username:                    cfg.GetString("RABBITMQ_USERNAME"),
		Password:                    cfg.GetString("RABBITMQ_PASSWORD"),
		Port:                        cfg.GetString("RABBITMQ_PORT"),
		Host:                        cfg.GetString("RABBITMQ_HOST"),
		Topic:                       cfg.GetString("RABBITMQ_TOPIC"),
		PathOriginalFile:            cfg.GetString("PATH_ORIGINAL_FILE"),
		PathCompressed:              cfg.GetString("PATH_COMPRESS_FILE"),
		SubPathOriginalInvtrypht:    cfg.GetString("SUB_PATH_ORIGINAL_INVTRYPHT"),
		SubPathOriginalAdjdmgpht:    cfg.GetString("SUB_PATH_ORIGINAL_ADJDMGPHT"),
		SubPathCompressionInvtrypht: cfg.GetString("SUB_PATH_COMPRESS_INVTRYPHT"),
		SubPathCompressionAdjdmgpht: cfg.GetString("SUB_PATH_COMPRESS_ADJDMGPHT"),
	}

	discordCfg := config.Discord{
		Url: cfg.GetString("URL_BOT_DISCORD"),
	}

	serverCfg := config.Server{
		Name: cfg.GetString("APP_ENV"),
	}

	imageSettingCfg := config.ImageSetting{
		Quality: cfg.GetInt("QUALITY_COMPRESS"),
	}

	compressed.ConsumerProcessing(logger, rabbitCfg, discordCfg, serverCfg, imageSettingCfg)
}
