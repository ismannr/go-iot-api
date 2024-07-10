package controller

import (
	"gin-crud/config"
	"gin-crud/service"
	"github.com/gin-gonic/gin"
)

func UserController(r *gin.Engine) {
	r.GET("/user", config.AuthFilter, service.GetUserData)
	r.PUT("/user", config.AuthFilter, service.UpdateData)

	r.GET("/devices", config.AuthFilter, service.GetAllUserDevices)

	r.GET("/device/:id", config.AuthFilter, service.GetDeviceById)
	r.PUT("/device/:id", config.AuthFilter, service.UpdateDeviceName)

	r.POST("/device/register/:id", config.AuthFilter, service.RegisterDeviceById)
	r.DELETE("/device/delete/:id", config.AuthFilter, service.DeleteDeviceById)

	r.POST("/device/monitor-date-time", config.AuthFilter, service.GetMonitoringData)

	r.POST("/group/create", config.AuthFilter, service.CreateDeviceGroup)
	r.GET("/group", config.AuthFilter, service.GetAllGroup)
	r.GET("/group/:id", config.AuthFilter, service.GetGroupById)
	r.DELETE("/group/:id", config.AuthFilter, service.DeleteGroupById)

	r.PUT("/device/to-group/:id", config.AuthFilter, service.AddDeviceToGroup)
	r.DELETE("/device/to-group/:id", config.AuthFilter, service.RemoveDeviceFromGroup)

}
