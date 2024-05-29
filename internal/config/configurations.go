package config

import (
	"fmt"
	"image-compressions/pkg/rabbitmq"
)

func Load() (*Configurations, error) {
	v, err := GlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	cfg := Configurations{
		RabbitMq: rabbitmq.Config{
			Username: v.GetString("RABBITMQ_USERNAME"),
			Password: v.GetString("RABBITMQ_PASSWORD"),
			Port:     v.GetString("RABBITMQ_PORT"),
			Host:     v.GetString("RABBITMQ_HOST"),
			Topic:    v.GetString("RABBITMQ_TOPIC"),
		},
		Discord: Discord{
			Url: v.GetString("URL_BOT_DISCORD"),
		},
		Server: Server{
			Name: v.GetString("APP_ENV"),
		},
		ImageSetting: ImageSetting{
			Quality:                     v.GetInt("QUALITY_COMPRESS"),
			PathOriginalFile:            v.GetString("PATH_ORIGINAL_FILE"),
			PathCompressed:              v.GetString("PATH_COMPRESS_FILE"),
			SubPathOriginalInvtrypht:    v.GetString("SUB_PATH_ORIGINAL_INVTRYPHT"),
			SubPathOriginalAdjdmgpht:    v.GetString("SUB_PATH_ORIGINAL_ADJDMGPHT"),
			SubPathCompressionInvtrypht: v.GetString("SUB_PATH_COMPRESS_INVTRYPHT"),
			SubPathCompressionAdjdmgpht: v.GetString("SUB_PATH_COMPRESS_ADJDMGPHT"),
		},
		Logger: Logger{
			Level: v.GetString("LOG_LEVEL"),
		},
	}

	return &cfg, err
}

type Configurations struct {
	RabbitMq     rabbitmq.Config
	Discord      Discord
	Server       Server
	ImageSetting ImageSetting
	Logger       Logger
}

type Discord struct {
	Url string
}

type Server struct {
	Name string
}

type ImageSetting struct {
	Quality                     int
	PathOriginalFile            string
	PathCompressed              string
	SubPathOriginalInvtrypht    string
	SubPathOriginalAdjdmgpht    string
	SubPathCompressionInvtrypht string
	SubPathCompressionAdjdmgpht string
}

type Logger struct {
	Level string
}
