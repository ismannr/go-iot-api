package controller

import (
	"gin-crud/config"
	"gin-crud/service"
	"github.com/gin-gonic/gin"
)

func GuestController(r *gin.Engine) {
	r.POST("/sign-up", service.UserRegister)

	r.POST("/login", service.Login)
	r.GET("/logout", config.AuthFilter, service.Logout)

	r.POST("/forgot-password", service.RecoveryPassword)
	r.POST("/reset-password/:token", config.RecoveryAuthFilter, service.ResetPassword)
}
