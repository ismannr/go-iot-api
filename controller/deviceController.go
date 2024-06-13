package controller

import (
	"gin-crud/service"
	"github.com/gin-gonic/gin"
)

func DeviceController(r *gin.Engine) {
	r.PUT("/device-gateway", service.ReceiveAndSaveData)
}
