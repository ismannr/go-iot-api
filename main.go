package main

import (
	"gin-crud/controller"
	"gin-crud/initializers"
	"gin-crud/service"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.DatabaseInit()
}

func main() {
	r := gin.Default()
	controller.UserController(r)
	controller.GuestController(r)
	controller.AdminController(r)
	controller.DeviceController(r)
	go service.TokenExpirationCheckAndUpdateScheduler()
	go service.ClearDeviceDataScheduler()
	r.Run()
}
