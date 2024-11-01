package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadDbEnv(envFile string) {
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatal("Error loading .env file") // Exit on error
	}
}
