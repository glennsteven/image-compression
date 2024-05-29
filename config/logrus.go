package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"image-compressions/consts"
)

func NewLogger(cfg Logger) (*logrus.Logger, error) {
	logger := logrus.New()

	lv, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("cannot parrse level: %s: %w", cfg.Level, err)
	}

	logger.SetLevel(lv)

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: consts.LayoutDateTimeFormat,
		ForceColors:     true,
		DisableColors:   false,
		PadLevelText:    true,
	}
	logger.SetFormatter(formatter)

	logger.WithField("level", lv.String()).Info("Logger initialized")
	return logger, nil
}
