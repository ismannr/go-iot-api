package service

import (
	"errors"
	"fmt"
	"gin-crud/initializers"
	models "gin-crud/models"
	"gin-crud/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

func generateToken(user models.SystemData, c *gin.Context) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	switch user.Role {
	case models.RoleBinusian, models.RoleUMKM:
		claims["sub"] = user.ID
		claims["exp"] = time.Now().Add(time.Hour).Unix()
		claims["role"] = string(user.Role)
	default:
		response.GlobalResponse(c, "Invalid user type", http.StatusInternalServerError, nil)
		return
	}

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		response.GlobalResponse(c, "Invalid token creation", http.StatusInternalServerError, nil)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, int(time.Hour.Seconds()), "", "", false, true)

	if user.TokenID == nil {
		tokenUser := &models.Token{
			ID:           uuid.New(),
			Token:        tokenString,
			TokenExpired: false,
			TokenExpiry:  time.Now().Add(time.Hour),
		}
		user.Token = tokenUser
	} else {
		var tokenId models.Token
		if result := initializers.DB.First(&tokenId, user.TokenID).Error; err != nil {
			log.Println(result.Error())
			response.GlobalResponse(c, "Failed to retrieve token data", http.StatusInternalServerError, nil)
			return
		}
		tokenId.Token = tokenString
		tokenId.TokenExpired = false
		tokenId.TokenExpiry = time.Now().Add(time.Hour)

		user.Token = &tokenId
	}

	if err := initializers.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&user).Error; err != nil {
		response.GlobalResponse(c, "Failed to update participant data", http.StatusInternalServerError, nil)
		return
	}
	response.GlobalResponse(c, "Token generated", 200, nil)
}

func confirmationToken(email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = email
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", err
	}
	endpoint := os.Getenv("CONFIRMATION_ENDPOINT")
	url := fmt.Sprintf("%s%s", endpoint, tokenString)
	return url, nil
}

func getUserByAuth(c *gin.Context) (interface{}, error) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		response.GlobalResponse(c, "Authorization token not provided", http.StatusUnauthorized, nil)
		return nil, errors.New("authorization token not provided")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub := claims["sub"].(string)
		userID, err := uuid.Parse(sub)
		if err != nil {
			return nil, errors.New("failed parsing user ID")
		}
		if claims["role"] == string(models.RoleBinusian) {
			var mentor models.BinusianData
			if err := initializers.DB.Preload("SystemData").First(&mentor, "system_data_id = ?", userID).Error; err == nil {
				return &mentor, nil
			}
		} else if claims["role"] == string(models.RoleUMKM) {
			var participant models.UmkmData
			if err := initializers.DB.Preload("SystemData").First(&participant, "system_data_id = ?", userID).Error; err == nil {
				return &participant, nil
			}
		}
		return nil, errors.New("user not found")
	}
	return nil, errors.New("invalid token")
}

func Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", -1, "", "", false, true)
	user, err := getUserByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "Unauthorized user", 403, nil)
		return
	}

	var sysData models.SystemData
	switch user.(type) {
	case *models.UmkmData:
		participant := user.(*models.UmkmData)
		participant.SystemData.CurrentlyLogin = false
		if err := initializers.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&participant.SystemData).Error; err != nil {
			response.GlobalResponse(c, "Token Expired", 403, nil)
			return
		}
		if sysResult := initializers.DB.First(&sysData, participant.SystemDataID).Error; err != nil {
			log.Println(sysResult.Error())
			response.GlobalResponse(c, "Failed to retrieve system data", http.StatusInternalServerError, nil)
			return
		}
	case *models.BinusianData:
		mentor := user.(*models.BinusianData)
		mentor.SystemData.CurrentlyLogin = false

		if err := initializers.DB.Save(&mentor.SystemData).Error; err != nil {
			response.GlobalResponse(c, "Token Expired", 403, nil)
			return
		}
		if sysResult := initializers.DB.First(&sysData, mentor.SystemDataID).Error; err != nil {
			log.Println(sysResult.Error())
			response.GlobalResponse(c, "Failed to retrieve system data", http.StatusInternalServerError, nil)
			return
		}
	default:
		response.GlobalResponse(c, "Invalid user type", 500, nil)
		return
	}
	var token models.Token
	if tokenResult := initializers.DB.First(&token, sysData.TokenID).Error; err != nil {
		log.Println(tokenResult.Error())
		response.GlobalResponse(c, "Failed to retrieve token data", http.StatusInternalServerError, nil)
		return
	}
	if err := initializers.DB.Unscoped().Delete(&token).Error; err != nil {
		log.Println(err.Error())
		response.GlobalResponse(c, "Failed to invalidate token", 500, nil)
		return
	}
	response.GlobalResponse(c, "Logout successful", 200, nil)
}
