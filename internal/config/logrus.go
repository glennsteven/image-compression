package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func NewLogger(cfg Logger) (*logrus.Logger, error) {
	lv, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("cannot parse level: %s, %w ", cfg.Level, err)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.DateTime,
		ForceColors:     true,
		DisableColors:   false,
		PadLevelText:    true,
	}

	logger := logrus.New()
	logger.SetLevel(lv)
	logger.SetFormatter(formatter)
	logger.WithField("level", lv).Info("Logger initialized")

	return logger, nil
}
