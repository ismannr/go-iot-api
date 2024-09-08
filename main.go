package main

import (
	"fmt"
	"gin-crud/controller"
	"gin-crud/initializers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

var logFile *os.File

func generateLogFilename() string {
	now := time.Now()
	return fmt.Sprintf("server_%s.log", now.Format("2006-01-02"))
}

func setupLogFile() {
	var err error
	logFileName := generateLogFilename()
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
}

func rotateLogFile() {
	for range time.NewTicker(3 * 30 * 24 * time.Hour).C { // 3 bulan
		logFile.Close()
		setupLogFile()
		log.Println("Log file rotated")
	}
}

func main() {
	setupLogFile()
	defer logFile.Close()

	go rotateLogFile()

	initializers.LoadEnvVariables()
	initializers.DatabaseInit()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://imon.andamantau.com"},
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

	//go service.TokenExpirationCheckAndUpdateScheduler()
	//go service.ClearDeviceDataScheduler()

	go func() {
		if err := r.Run(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	select {}
}
