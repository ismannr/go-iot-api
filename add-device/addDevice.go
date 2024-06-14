package main

import (
	"gin-crud/initializers"
	"gin-crud/models"
	"github.com/google/uuid"
	"log"
)

func main() {
	initializers.LoadEnvVariables()
	initializers.DatabaseInit()

	device := models.Device{
		ID: uuid.New(),
	}

	if err := models.CreateDevice(&device); err != nil {
		log.Println("failed creating new device")
		return
	}

	log.Println("successfully adding a device")
}
