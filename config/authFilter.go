package config

import (
	"errors"
	"fmt"
	"gin-crud/initializers"
	model "gin-crud/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

func AuthFilter(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var tokenDb model.Token
	if err := initializers.DB.First(&tokenDb, "token = ?", tokenString).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("Token not found in the database")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Println("Error retrieving token data:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok || float64(time.Now().Unix()) > exp {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		subUUID, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var user model.UmkmData
		if err := initializers.DB.Where("system_data_id = ?", subUUID).First(&user).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)

	}
}

func AdminAuthFilter(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var tokenDb model.Token
	if err := initializers.DB.First(&tokenDb, "token = ?", tokenString).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("Token not found in the database")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Println("Error retrieving token data:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok || float64(time.Now().Unix()) > exp {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		subUUID, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var user model.UmkmData
		if err := initializers.DB.Preload("SystemData").Where("system_data_id = ?", subUUID).First(&user).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if user.SystemData.Level != model.LevelAdmin {
			c.AbortWithStatus(403)
			return
		}
		c.Set("user", user)
		c.Next()
		return
	}

	c.AbortWithStatus(http.StatusUnauthorized)
}

func RecoveryAuthFilter(c *gin.Context) {
	var recoveryToken model.PasswordRecoveryToken
	accessToken := c.Param("token")

	if accessToken == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := initializers.DB.Where("rand_token_access = ?", accessToken).First(&recoveryToken).Error; err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	expiry := time.Minute * 5
	if recoveryToken.CreatedAt.Add(expiry).Before(time.Now()) {
		initializers.DB.Unscoped().Delete(&recoveryToken)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Set("accessToken", accessToken)
	c.Next()
}
