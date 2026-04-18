package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var (
	logger *log.Logger
	cfg    *Config
)

func init() {

	var confName string
	env := os.Getenv("ENV")

	if env == "container" {
		confName = "container"
	} else {
		confName = "local"
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "."
	}

	// init log system
	logger = log.Default()
	logger.SetOutput(os.Stdout)

	viper.SetConfigName(confName) // Name of the config file (without extension)
	viper.SetConfigType("yaml")   // Config file type
	viper.AddConfigPath(configPath)
	viper.AddConfigPath("config")

	// Read the config file
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Printf("Warning: Config file not found, using default values: %v", err)
		} else {
			logger.Fatalf("Error reading config file: %v", err)
		}
	}

	// Initialize variables by unmarshaling into the struct
	cfg = new(Config)
	err = viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}
}

func GetLogger() *log.Logger {
	return logger
}

func GetConfig() *Config {
	return cfg
}
