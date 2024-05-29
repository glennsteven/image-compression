package config

import (
	"github.com/sirupsen/logrus"
	"image-compressions/consts"
)

func NewLogger(cfg Logger) *logrus.Logger {
	logger := logrus.New()

	level := logrus.Level(cfg.Level)
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
