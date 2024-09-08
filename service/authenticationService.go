package service

import (
	"errors"
	"fmt"
	"gin-crud/initializers"
	model "gin-crud/models"
	"gin-crud/request"
	"gin-crud/response"
	"gin-crud/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

func Login(c *gin.Context) {
	var systemData model.SystemData
	var req struct {
		Email    string
		Password string
	}
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Failed to retrieve systemData request", http.StatusBadRequest, nil)
		return
	}
	result := initializers.DB.First(&systemData, "email = ?", req.Email)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.GlobalResponse(c, "Invalid email or password", http.StatusBadRequest, nil)
		return
	} else if result.Error != nil {
		log.Println("Error :", result.Error)
		response.GlobalResponse(c, "Internal server error", http.StatusInternalServerError, nil)
		return
	}
	if utils.HashIsMatched(systemData.Password, req.Password) == false {
		response.GlobalResponse(c, "Invalid email or password", http.StatusBadRequest, nil)
		return
	}

	systemData.CurrentlyLogin = true
	systemData.LastLogin = time.Now()
	if err := initializers.DB.Updates(&systemData).Error; err != nil {
		response.GlobalResponse(c, "Failed to update login status", http.StatusInternalServerError, nil)
		return
	}
	generateToken(systemData, c)
}

func UserRegister(c *gin.Context) {
	var req request.UmkmRequest
	var s strings.Builder
	var userDB model.UmkmData
	var isSatisfied bool = true

	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Failed to retrieve user request", http.StatusBadRequest, nil)
		return
	}
	if len(req.Name) < 3 {
		s.WriteString("Name, ")
		isSatisfied = false
	}
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		s.WriteString("Email (wrong format), ")
		isSatisfied = false
	} else if exist := initializers.DB.Where("email = ?", req.Email).First(&userDB).Error; exist == nil {
		s.WriteString("Email already exist, ")
		isSatisfied = false
	}
	if strings.ToLower(req.Gender) != "male" && strings.ToLower(req.Gender) != "female" {
		s.WriteString("Gender, ")
		isSatisfied = false
	}
	dob, err := utils.ParseDate(req.Dob)
	if err != nil {
		s.WriteString("DOB (wrong date format), ")
		isSatisfied = false
	} else {
		if !utils.IsAdult(req.Dob) {
			s.WriteString("DOB (Must be over 17), ")
			isSatisfied = false
		}
	}

	if len(req.Password) < 8 {
		s.WriteString("Password min. 8 char, ")
		isSatisfied = false
	} else {
		uppercaseRegex := regexp.MustCompile(`[A-Z]`)
		numberRegex := regexp.MustCompile(`[0-9]`)
		if !uppercaseRegex.MatchString(req.Password) || !numberRegex.MatchString(req.Password) {
			s.WriteString("Password (Must contain at least one uppercase letter and one number), ")
			isSatisfied = false
		} else if req.ConfirmPass != req.Password {
			s.WriteString("Confirmation Password doesn't match, ")
			isSatisfied = false
		}
	}
	if exist := initializers.DB.Where("phone = ?", req.PhoneNumber).First(&userDB).Error; exist == nil {
		s.WriteString("Phone number already exist, ")
		isSatisfied = false
	}

	if req.Province == "" {
		s.WriteString("Province cannot be empty, ")
		isSatisfied = false
	}

	if req.City == "" {
		s.WriteString("City cannot be empty, ")
		isSatisfied = false
	}

	if req.Address == "" {
		s.WriteString("Address cannot be empty, ")
		isSatisfied = false
	}

	if req.BusinessName == "" {
		s.WriteString("Address cannot be empty, ")
		isSatisfied = false
	}

	if req.BusinessDesc == "" {
		s.WriteString("Address cannot be empty, ")
		isSatisfied = false
	}

	if !isSatisfied {
		message := "User data requirements not satisfied: " + s.String()
		response.GlobalResponse(c, message, http.StatusBadRequest, nil)
		return
	}

	password, err := utils.HashEncoder(req.Password)
	if err != nil {
		message := "Error encoding the password"
		response.GlobalResponse(c, message, http.StatusBadRequest, nil)
		return
	}
	userId := uuid.New()
	systemUser := model.SystemData{
		ID:       userId,
		Email:    req.Email,
		Password: password,
		Role:     model.RoleUMKM,
		Level:    model.LevelUser,
	}
	user := model.UmkmData{
		ID:           userId,
		Name:         req.Name,
		Gender:       strings.ToUpper(req.Gender),
		Dob:          dob,
		Email:        req.Email,
		Address:      req.Address,
		Province:     req.Province,
		City:         req.City,
		BusinessName: req.BusinessName,
		BusinessDesc: req.BusinessDesc,
		Phone:        req.PhoneNumber,
		SystemDataID: &systemUser.ID,
		SystemData:   &systemUser,
	}

	result := initializers.DB.Create(&user)
	if result.Error != nil {
		response.GlobalResponse(c, "Failed to save user data", http.StatusBadRequest, nil)
		return
	}
	initializers.DB.Save(&user)
	if err != nil {
		response.GlobalResponse(c, "Failed to generate confirmation token", http.StatusInternalServerError, nil)
	}
	r, err := RegistrationMail(req.Email, req.Name)
	if err != nil {
		log.Println("Failed to send registration confirmation: " + r)
		response.GlobalResponse(c, "Failed to send registration confirmation", http.StatusOK, nil)
	}
	respString := fmt.Sprintf("Your account has been created. %s, please check your email.", r)
	response.GlobalResponse(c, respString, http.StatusOK, nil)
}

func RecoveryPassword(c *gin.Context) {
	var req request.RecoveryRequest
	var user model.UmkmData
	var SysData *model.SystemData
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Failed to retrieve user request", http.StatusBadRequest, nil)
		return
	}

	if len(req.Email) == 0 {
		response.GlobalResponse(c, "Email cannot be empty!", http.StatusBadRequest, nil)
		return
	}

	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		response.GlobalResponse(c, "Email doesnt exist!", http.StatusBadRequest, nil)
		return
	}

	if err := initializers.DB.Preload("RecoveryToken").Where("id = ?", user.SystemDataID).First(&SysData).Error; err != nil {
		response.GlobalResponse(c, "Email doesnt exist on system data!", http.StatusBadRequest, nil)
		return
	}

	RandAccess := generateRandomString(64)
	url := "https://imon.andamantau.com/inputnewpassword/" + RandAccess

	if SysData.RecoveryToken != nil {
		SysData.RecoveryToken.RandTokenAccess = RandAccess
	} else {
		SysData.RecoveryToken = &model.PasswordRecoveryToken{
			ID:              uuid.New(),
			RandTokenAccess: RandAccess,
		}
	}
	_, err := ForgotPasswordMail(req.Email, user.Name, url)
	if err != nil {
		log.Println("Failed to send mail: " + err.Error())
		response.GlobalResponse(c, "Failed to send mail", http.StatusInternalServerError, nil)
		return
	}
	if err := initializers.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&SysData).Error; err != nil {
		response.GlobalResponse(c, "Failed to save recovery token", http.StatusBadRequest, nil)
		log.Println("Failed to save recovery token: " + err.Error())
		return
	}
	response.GlobalResponse(c, "Successfully sending password recovery mail", http.StatusOK, nil)
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func ResetPassword(c *gin.Context) {
	var user *model.SystemData
	var recoveryToken *model.PasswordRecoveryToken
	var req request.ResetPasswordRequest
	accessToken, exist := c.Get("accessToken")

	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Failed to retrieve user request", http.StatusBadRequest, nil)
		return
	}

	if !exist {
		response.GlobalResponse(c, "Token doesn't exist", http.StatusBadRequest, nil)
	}

	if err := initializers.DB.Where("rand_token_access = ?", accessToken).First(&recoveryToken).Error; err != nil {
		response.GlobalResponse(c, "Can't find the user", http.StatusInternalServerError, nil)
		return
	}

	if err := initializers.DB.Where("recovery_token_id = ?", recoveryToken.ID).First(&user).Error; err != nil {
		response.GlobalResponse(c, "Can't find the user", http.StatusInternalServerError, nil)
		return
	}

	if req.NewPassword != req.PasswordConfirmation {
		response.GlobalResponse(c, "Password confirmation doesn't match", http.StatusBadRequest, nil)
		return
	}

	password, err := utils.HashEncoder(req.NewPassword)
	if err != nil {
		response.GlobalResponse(c, "Can't encode the password", http.StatusInternalServerError, nil)
		return
	}
	user.Password = password
	initializers.DB.Save(&user)
	initializers.DB.Unscoped().Delete(&recoveryToken)
	response.GlobalResponse(c, "Successfully updating user password", http.StatusOK, nil)
}
