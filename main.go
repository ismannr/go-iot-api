package main

import (
	"gin-crud/controller"
	"gin-crud/initializers"
	"gin-crud/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	initializers.LoadEnvVariables()
	initializers.DatabaseInit()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // pake url frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
