package controller

import (
	"gin-crud/service"
	"github.com/gin-gonic/gin"
)

func DeviceController(r *gin.Engine) {
	r.GET("/device-gateway/:id/:oxygen_level/:water_temp/:ec_level/:ph_level/:time_stamp", service.ReceiveAndSaveData)
}
