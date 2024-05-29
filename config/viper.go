package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func GlobalConfig() (*viper.Viper, error) {
	config := viper.New()

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Construct the absolute path to the .env file.
	envFilePath := filepath.Join(currentDir, ".env")

	// Set the absolute path to the .env file in the Viper configuration.
	config.SetConfigFile(envFilePath)

	config.SetDefault("APP_PORT", 8000)

	err = config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}
