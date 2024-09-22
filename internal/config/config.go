package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	ServerPort  string
	EsHost      string
	EsUser      string
	EsPassword  string
}

func Init() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf(".env not found, please make sure .env file exists")
	}

	// init, check data and set default value
	config := &Config{}

	if environment := os.Getenv("ENVIRONMENT"); len(environment) > 0 {
		config.Environment = environment
	}

	if serverPort := os.Getenv("SERVER_PORT"); len(serverPort) > 0 {
		config.ServerPort = serverPort
	}

	if esHost := os.Getenv("ES_HOST"); len(esHost) > 0 {
		config.EsHost = esHost
	}

	if esUser := os.Getenv("ES_USER"); len(esUser) > 0 {
		config.EsUser = esUser
	}

	if esPass := os.Getenv("ES_PASSWORD"); len(esPass) > 0 {
		config.EsPassword = esPass
	}

	return config
}
