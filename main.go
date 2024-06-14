package main

import (
	"gin-crud/controller"
	"gin-crud/initializers"
	"gin-crud/service"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	initializers.LoadEnvVariables()
	initializers.DatabaseInit()

	r := gin.Default()
	controller.UserController(r)
	controller.GuestController(r)
	controller.AdminController(r)
	controller.DeviceController(r)
	go service.TokenExpirationCheckAndUpdateScheduler()
	go service.ClearDeviceDataScheduler()
	go func() {
		if err := r.Run(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	select {}
}
