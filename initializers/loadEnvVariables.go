package initializers

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadEnvVariables() {
	err := godotenv.Load() //loading custom port from env
	if err != nil {
		log.Fatal("Error loading .env file!")
	}
}
