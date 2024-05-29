package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"image-compressions/consts"
)

func NewLogger(viper *viper.Viper) *logrus.Logger {
	logger := logrus.New()

	level := logrus.Level(viper.GetInt32("LOG_LEVEL"))
	logger.SetLevel(level)

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: consts.LayoutDateTimeFormat,
		ForceColors:     true,
		DisableColors:   false,
		PadLevelText:    true,
	}
	logger.SetFormatter(formatter)

	logger.WithField("level", level.String()).Info("Logger initialized")
	return logger
}
