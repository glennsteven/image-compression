package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

func GlobalConfig() (*viper.Viper, error) {
	// Set the absolute path to the .env file in the Viper configuration.
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	envFilePath := filepath.Join(currentDir, ".env")

	config := viper.New()
	config.SetConfigFile(envFilePath)
	config.SetDefault("APP_PORT", 8000)

	err = config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}
