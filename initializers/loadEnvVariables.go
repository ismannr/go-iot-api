package initializers

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func LoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		message := fmt.Sprintf("Error loading .env file! Error: %s", err.Error())
		log.Fatal(message)
	}
}
