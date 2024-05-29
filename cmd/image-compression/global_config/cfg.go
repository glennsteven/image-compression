package global_config

import (
	"github.com/spf13/viper"
	"image-compressions/config"
)

func Config(cfg *viper.Viper) config.Configurations {
	return config.Configurations{
		RabbitMq: config.RabbitMq{
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
		},
		Discord: config.Discord{
			Url: cfg.GetString("URL_BOT_DISCORD"),
		},
		Server: config.Server{
			Name: cfg.GetString("APP_ENV"),
		},
		ImageSetting: config.ImageSetting{
			Quality: cfg.GetInt("QUALITY_COMPRESS"),
		},
	}

}
