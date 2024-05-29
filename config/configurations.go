package config

import (
	"fmt"
)

type RabbitMq struct {
	Username                    string
	Password                    string
	Port                        string
	Host                        string
	Topic                       string
	PathOriginalFile            string
	PathCompressed              string
	SubPathOriginalInvtrypht    string
	SubPathOriginalAdjdmgpht    string
	SubPathCompressionInvtrypht string
	SubPathCompressionAdjdmgpht string
}

type Discord struct {
	Url string
}

type Server struct {
	Name string
}

type ImageSetting struct {
	Quality int
}

type Config struct {
	RabbitMQ     RabbitMq
	Discord      Discord
	Server       Server
	ImageSetting ImageSetting
	Logger       Logger
}

type Logger struct {
	Level int32
}

func Load() (*Config, error) {
	v, err := GlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg := Config{
		RabbitMQ: RabbitMq{
			Username:                    v.GetString("RABBITMQ_USERNAME"),
			Password:                    v.GetString("RABBITMQ_PASSWORD"),
			Port:                        v.GetString("RABBITMQ_PORT"),
			Host:                        v.GetString("RABBITMQ_HOST"),
			Topic:                       v.GetString("RABBITMQ_TOPIC"),
			PathOriginalFile:            v.GetString("PATH_ORIGINAL_FILE"),
			PathCompressed:              v.GetString("PATH_COMPRESS_FILE"),
			SubPathOriginalInvtrypht:    v.GetString("SUB_PATH_ORIGINAL_INVTRYPHT"),
			SubPathOriginalAdjdmgpht:    v.GetString("SUB_PATH_ORIGINAL_ADJDMGPHT"),
			SubPathCompressionInvtrypht: v.GetString("SUB_PATH_COMPRESS_INVTRYPHT"),
			SubPathCompressionAdjdmgpht: v.GetString("SUB_PATH_COMPRESS_ADJDMGPHT"),
		},
		Discord: Discord{
			Url: v.GetString("URL_BOT_DISCORD"),
		},
		Server: Server{
			Name: v.GetString("APP_ENV"),
		},
		ImageSetting: ImageSetting{
			Quality: v.GetInt("QUALITY_COMPRESS"),
		},
		Logger: Logger{
			v.GetInt32("LOG_LEVEL"),
		},
	}

	return &cfg, nil
}
